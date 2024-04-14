package udp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"golang.org/x/net/ipv4"
)

type Tcp struct {
	conn          net.Conn
	r             *bufio.Reader
	lengthByte    [2]byte
	rBuffer       sync.Pool
	sBuffer       sync.Pool
	SendChan      chan []*Udpdata
	udpstructlist []*Udpdata
	SendListChan  chan *TcpSendListStruct
}

type TcpSendListStruct struct {
	header      []byte
	listData    [][]byte
	messagePool []ipv4.Message
	dataPool    [][]byte
}

func NewTcp(conn net.Conn, server bool) (*Tcp, error) {
	tcpconn, ok := conn.(*net.TCPConn)
	if ok {
		tcpconn.SetNoDelay(true)
		// tcpconn.SetKeepAlive(true)
		// tcpconn.SetReadBuffer(3 << 21)
		// tcpconn.SetWriteBuffer(3 << 21)
	}

	resp := &Tcp{conn: conn, udpstructlist: make([]*Udpdata, 0, 35),
		SendListChan: make(chan *TcpSendListStruct, 35),
	}
	if server {
		ok, err := resp.Validate()
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("Validate is faild")
		}
	} else {
		resp.Register()
	}
	resp.rBuffer.New = func() any {
		return make([]byte, 65535)
	}
	resp.sBuffer.New = func() any {
		return make([]byte, 65535)
	}
	resp.SendChan = make(chan []*Udpdata, 35)
	resp.r = bufio.NewReader(resp.conn)
	return resp, nil
}
func (t *Tcp) Validate() (bool, error) {
	b := make([]byte, 10)
	check := []byte("okokokokok")
	_, err := t.conn.Read(b)

	if err != nil {
		return false, err
	}
	t.conn.Write(check)

	return bytes.Equal(b, check), nil
}
func (t *Tcp) Register() (bool, error) {
	b := make([]byte, 10)
	check := []byte("okokokokok")
	_, err := t.conn.Write(check)
	if err != nil {
		return false, err
	}
	_, err = t.conn.Read(b)
	if err != nil {
		return false, err
	}
	return bytes.Equal(b, check), nil
}
func (t *Tcp) SendUseChan() {
	for v := range t.SendChan {
		t.StructToWrite(v)
	}
}

func (t *Tcp) ReadtoStruct() ([]*Udpdata, error) {
	// data := t.rBuffer
	structList := t.udpstructlist
	_, err := io.ReadFull(t.r, t.lengthByte[:])
	if err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint16(t.lengthByte[:])
	data := make([]byte, length)
	n, err := io.ReadFull(t.r, data)
	if err != nil || n != int(length) {
		return nil, err
	}
	for len(data) > 2 {
		structLength := binary.LittleEndian.Uint16(data[:2])
		if len(data) >= int(structLength) {
			tempStruct := Udpdata{}
			tempStruct.Unmarshal(data[:structLength])
			structList = append(structList, &tempStruct)
			data = data[structLength:]

		} else {
			fmt.Println("data[:length]: ", length)
			log.Println("structLength: ", structLength, "len(data): ", len(data))
			return nil, fmt.Errorf("struct decode error")
		}

	}
	if len(data) > 0 {
		log.Fatal("len(data)!=0")
	}
	return structList, nil
}
func (t *Tcp) Close() {
	close(t.SendChan)
	close(t.SendListChan)
	return
}
func (t *Tcp) StructToWrite(list []*Udpdata) (n int, err error) {
	var dataLenght int
	data := t.sBuffer.Get().([]byte)[:0]
	defer t.sBuffer.Put(data)
	for _, v := range list {
		dataLenght += int(v.Length)
	}
	data = binary.LittleEndian.AppendUint16(data, uint16(dataLenght))
	for _, v := range list {
		data = append(data, v.Marshal()...)
	}
	return t.conn.Write(data)
}

func (t *Tcp) SendUseChanList() {
	for v := range t.SendListChan {
		t.TcpListToWrite(v.header, v.listData)
		DataListPool.Put(v.dataPool)
		MessagePool.Put(v.messagePool)
	}
}

func (t *Tcp) TcpListToWrite(srcHeader []byte, list [][]byte) (n int, err error) {
	var dataLenght int
	data := t.sBuffer.Get().([]byte)[:0]
	defer t.sBuffer.Put(data)
	for _, v := range list {
		dataLenght = dataLenght + len(v) + 8
	}
	data = binary.LittleEndian.AppendUint16(data, uint16(dataLenght))
	for _, v := range list {
		data = binary.LittleEndian.AppendUint16(data, uint16(len(v)+8))
		data = append(data, srcHeader...)
		data = append(data, v...)
	}
	return t.conn.Write(data)
}
