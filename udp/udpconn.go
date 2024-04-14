package udp

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"syscall"
	"time"

	"golang.org/x/net/ipv4"
)

const (
	UDP_SEGMENT = 103
)

type UdpConn struct {
	Conn             *ipv4.PacketConn
	Mtu              int
	BatchSize        int
	sendMsgs         []ipv4.Message
	readMsgs         []ipv4.Message
	DataChan         chan *Udpdata
	udpstructlist    []*Udpdata
	srcIPandPortByte [6]byte
}

type SendChanStruct struct {
	*UdpConn
}

func NewUdpListen(localIP *net.UDPAddr, mtu, batchSize int) *UdpConn {
	udpconn, err := net.ListenUDP("udp", localIP)
	if err != nil {
		fmt.Println(err.Error())
	}
	packetConn := ipv4.NewPacketConn(udpconn)
	// msgs := make([]ipv4.Message, 10)
	respon := &UdpConn{
		Conn:          packetConn,
		Mtu:           mtu,
		DataChan:      make(chan *Udpdata, 35),
		udpstructlist: make([]*Udpdata, 0, 35),
	}
	smsgs := make([]ipv4.Message, batchSize)
	rmsgs := make([]ipv4.Message, batchSize)
	for i := 0; i < batchSize; i++ {
		rmsgs[i].Buffers = [][]byte{make([]byte, 1500)}

	}
	respon.readMsgs = rmsgs
	respon.sendMsgs = smsgs
	return respon
}

func (c *UdpConn) ClientRendToStruct() ([]*Udpdata, error) {
	tmpUdpDataList := c.udpstructlist
	readMsgs := c.readMsgs
	n, err := c.Conn.ReadBatch(readMsgs, 0)
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		Length := readMsgs[i].N
		if Length == 0 {
			continue
		}
		udpStruct, err := NewUdpData(readMsgs[i].Addr, readMsgs[i].Buffers[0][:Length])
		if err != nil {
			return nil, err
		}
		tmpUdpDataList = append(tmpUdpDataList, udpStruct)
	}
	return tmpUdpDataList, nil
}

func (c *UdpConn) Close() {
	c.Conn.Close()
	close(c.DataChan)
}

func (c *UdpConn) ClientNormalSend(data []*Udpdata) {
	for _, v := range data {
		c.Conn.WriteTo(v.Data, nil, &net.UDPAddr{
			IP:   v.DstIP,
			Port: int(v.DstPort),
		})
	}
}

func (c *UdpConn) ServerNormalSend(deadlineTime time.Duration) {
	deadliner := time.NewTimer(deadlineTime)
	for {
		select {
		case v := <-c.DataChan:
			c.Conn.WriteTo(v.Data, nil, Dst)
			deadliner.Reset(deadlineTime)
		case <-deadliner.C:
			deadliner.Stop()
			c.Close()
			delete(IdForchanMap, c.srcIPandPortByte)
			log.Println("The connection with ID %v was closed due to a timeout.", c.srcIPandPortByte)
			return
		}
	}

}

func (c *UdpConn) ServerRendToList(readMsgs []ipv4.Message, dataList [][]byte) (*TcpSendListStruct, error) {
	n, err := c.Conn.ReadBatch(readMsgs, 0)
	if err != nil {
		return nil, err
	}
	for _, v := range readMsgs[:n] {
		if v.N == 0 {
			continue
		}
		dataList = append(dataList, v.Buffers[0][:v.N])
	}
	return &TcpSendListStruct{
		c.srcIPandPortByte[:], dataList,
		readMsgs, dataList,
	}, nil
}

// func (c *UdpConn) ServerRendToStruct() ([]*Udpdata, error) {
// 	tmpUdpDataList := c.udpstructlist
// 	readMsgs := c.readMsgs
// 	n, err := c.Conn.ReadBatch(readMsgs, 0)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for i := 0; i < n; i++ {
// 		Length := readMsgs[i].N
// 		if Length <= 0 {
// 			continue
// 		}

// 		udpStruct, err := NewUdpData(c.src, readMsgs[i].Buffers[0][:Length])
// 		if err != nil {
// 			return nil, err
// 		}
// 		tmpUdpDataList = append(tmpUdpDataList, udpStruct)
// 	}
// 	return tmpUdpDataList, nil
// }

func NewUdpConn(srcIP *net.UDPAddr, mtu, batchSize int) *UdpConn {
	udpconn, err := net.DialUDP("udp", nil, Dst)
	if err != nil {
		fmt.Println(err.Error())
	}
	packetConn := ipv4.NewPacketConn(udpconn)
	// msgs := make([]ipv4.Message, 10)
	respon := &UdpConn{
		Conn: packetConn,
		// src:  srcIP,
		// dst:  targetIP,
		Mtu: mtu,

		DataChan: make(chan *Udpdata, 35),
		// messagePool: sync.Pool{
		// 	New: func() any {
		// 		temp := make([]ipv4.Message, batchSize)
		// 		for i := 0; i < batchSize; i++ {
		// 			temp[i].Buffers = [][]byte{make([]byte, 1500)}
		// 		}
		// 		return temp
		// 	},
		// },
	}

	smsgs := make([]ipv4.Message, 35)
	rmsgs := make([]ipv4.Message, batchSize)
	// udpAddr, err := net.ResolveUDPAddr(srcIP.Network(), respon.src.String())
	// if err != nil {
	// 	return nil
	// }
	binary.LittleEndian.PutUint16(respon.srcIPandPortByte[:2], srcIP.AddrPort().Port())
	copy(respon.srcIPandPortByte[2:6], srcIP.IP.To4())
	for i := 0; i < batchSize; i++ {
		rmsgs[i].Buffers = [][]byte{make([]byte, 1500)}
		rmsgs[i].OOB = make([]byte, 100)
	}
	respon.readMsgs = rmsgs
	respon.sendMsgs = smsgs
	return respon

}

func getCmsg(size int, b []byte) []byte {
	binary.LittleEndian.PutUint64(b[:8], 18)
	binary.LittleEndian.PutUint32(b[8:12], uint32(syscall.IPPROTO_UDP))
	binary.LittleEndian.PutUint32(b[12:16], uint32(UDP_SEGMENT))
	binary.LittleEndian.PutUint16(b[16:], uint16(size))
	return b
}

// func (c *UdpConn) GSOListenSend(data [][]byte, cmsg_OOB_szise int, addr net.Addr) (int, error) {
// 	if len(data) > len(c.sendMsgs) {
// 		fmt.Println("len(data)", len(data), len(c.sendMsgs))
// 		return 0, fmt.Errorf("data size is to lager")

// 	}
// 	for i, v := range data {
// 		c.sendMsgs[i].Buffers = [][]byte{v}
// 		if cmsg_OOB_szise > 0 {
// 			c.sendMsgs[i].OOB = getCmsg(cmsg_OOB_szise, make([]byte, 100))
// 		}
// 		c.sendMsgs[i].Addr = addr
// 	}
// 	n, err := c.Conn.WriteBatch(c.sendMsgs, 0)

// 	return n, err
// }
// func (c *UdpConn) GSOSend(data [][]byte, cmsg_OOB_szise int) (int, error) {
// 	if len(data) > len(c.sendMsgs) {
// 		return 0, fmt.Errorf("data size is to lager")
// 	}
// 	for i, v := range data {
// 		c.sendMsgs[i].Buffers = [][]byte{v}
// 		if cmsg_OOB_szise > 0 {
// 			c.sendMsgs[i].OOB = getCmsg(cmsg_OOB_szise, make([]byte, 100))
// 		}
// 	}
// 	n, err := c.Conn.WriteBatch(c.sendMsgs[:len(data)], 0)
// 	return n, err
// }

// func (c *UdpConn) Read(data [][]byte) (int, error) {
// 	n, err := c.Conn.ReadBatch(c.readMsgs, 0)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if len(data) >= n {
// 		for i := 0; i < n; i++ {
// 			udpPack, _ := NewUdpData(c.readMsgs[i].Addr, c.readMsgs[i].Buffers[0])
// 			data = append(data, udpPack.Marshal())
// 		}
// 	} else {
// 		return 0, fmt.Errorf("data Length too sort")
// 	}
// 	return n, nil
// }
