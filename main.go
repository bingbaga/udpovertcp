package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"gso/udp"
	"io"
	"log"
	"net"

	"sync"

	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"
)

func send() {
	targetIP := &net.UDPAddr{
		IP:   net.IPv4(192, 168, 1, 112),
		Port: 22222,
	}
	udpconn, err := net.DialUDP("udp", nil, targetIP)
	if err != nil {
		fmt.Println(err.Error())
	}
	
	fmt.Println(supportsUDPOffload(udpconn))
	packetConn := ipv4.NewPacketConn(udpconn)
	msgs := make([]ipv4.Message, 2)
	// var sizeOfGSOData = 2
	// var stickyControlSize = unix.CmsgSpace(unix.SizeofInet6Pktinfo)
	// var gsoControlSize = unix.CmsgSpace(sizeOfGSOData)
	for i := range msgs {
		//每一组buffers 有一个OOB 控制
		msgs[i].Buffers = [][]byte{make([]byte, 1600), make([]byte, 1000)}
		//OOB设置GSO 分割大小
	}
	for i := 0; i < 1; i++ {
		n, err := packetConn.WriteBatch(msgs, 0)
		if err != nil {
			fmt.Println(err.Error(), n)
		}
	}

}

func read() {
	localIP := &net.UDPAddr{
		IP:   net.IPv4(192, 168, 1, 23),
		Port: 22223,
	}
	var addr net.Addr
	udpconn := udp.NewUdpListen(localIP, 1400, 35)

	tcpConn, err := net.Dial("tcp", "aws.lbtest.top:33366")
	if err != nil {
		log.Fatal(err.Error())
	}
	tcp, _ := tcpConn.(*net.TCPConn)
	tcp.SetNoDelay(true)
	tcp.SetReadBuffer(3 * 1024 * 1024)
	tcp.SetWriteBuffer(3 * 1024 * 1024)


	var bytePool sync.Pool
	bytePool.New = func() any {
		return make([]byte, 65535)
	}
	go func() {
		defer tcpConn.Close()
		bconn := bufio.NewReader(tcpConn)
		for {
			data := bytePool.Get().([]byte)
			lengthByte := make([]byte, 2)
			_, err := bconn.Read(lengthByte)
			if err != nil {
				fmt.Println("Read error:", err)
				if err == io.EOF {
					return
				}
			}
			dataLength := binary.LittleEndian.Uint16(lengthByte)
			_, err = io.ReadFull(bconn, data[:dataLength])
			if err != nil {
				fmt.Println("ReadFull error:", err)
				if err == io.EOF {
					return
				}
			}

			remainingLen, datalist := handleData(data[:dataLength])
			bytePool.Put(data)
			if remainingLen>0{
				continue
			}
			if len(datalist) == 0 {
				fmt.Printf("binary.LittleEndian.Uint16(data[:2]): %v\n", binary.LittleEndian.Uint16(data[:2]))
				fmt.Println("bug bug bug", dataLength)
				fmt.Println(datalist)
				fmt.Println("remainingLen", remainingLen)
				continue
			}
		
			_, err = udpconn.GSOListenSend(datalist, 0, addr)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			
		}
	}()
	for {
		n, data,taddr,err := udpconn.RendToList()
		addr=taddr
		if err != nil {
			log.Fatal(err.Error())
		}		
		lengthByte:=make([]byte,2)
	
		binary.LittleEndian.PutUint16(lengthByte,uint16(n))
		
			_, err = tcpConn.Write(append(lengthByte, data...))		
	}

}
func handleData(data []byte) (int, [][]byte) {
	var dataList [][]byte // 动态地添加元素，而不是预分配
	for len(data) >= 2 {  // 确保有足够的数据来读取长度信息

		dataLength := binary.LittleEndian.Uint16(data[:2])
		
		totalLength := int(dataLength)+2  // 数据长度加上长度字段本身的2字节
		if len(data) >= totalLength {      // 确保有足够的数据来读取完整的数据段
			dataSegment := data[2:totalLength]
			dataList = append(dataList, dataSegment)
			data = data[totalLength:] // 移动到下一个数据段的开始位置
		} else {
			break // 剩余数据不足以构成一个完整的数据段
		}
	}
	return len(data), dataList // 返回未处理数据的长度和解析出的数据段列表
}

func main() {

	read()
}

func supportsUDPOffload(conn *net.UDPConn) (txOffload, rxOffload bool) {
	rc, err := conn.SyscallConn()
	if err != nil {
		return
	}
	err = rc.Control(func(fd uintptr) {
		_, errSyscall := unix.GetsockoptInt(int(fd), unix.IPPROTO_UDP, unix.UDP_SEGMENT)
		txOffload = errSyscall == nil
		opt, errSyscall := unix.GetsockoptInt(int(fd), unix.IPPROTO_UDP, unix.UDP_GRO)
		rxOffload = errSyscall == nil && opt == 1
	})
	if err != nil {
		return false, false
	}
	return txOffload, rxOffload
}
