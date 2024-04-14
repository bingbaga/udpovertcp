package udp

import (
	"net"
	"sync"

	"golang.org/x/net/ipv4"
)

var IdForchanMap map[[6]byte]chan *Udpdata
var MessagePool sync.Pool
var DataListPool sync.Pool
var Dst *net.UDPAddr

func init() {
	IdForchanMap = make(map[[6]byte]chan *Udpdata, 100)
	MessagePool.New = func() any {
		temp := make([]ipv4.Message, 35)
		for i := 0; i < 35; i++ {
			temp[i].Buffers = [][]byte{make([]byte, 1500)}
		}
		return temp
	}
	DataListPool.New = func() any {
		temp := make([][]byte, 35)
		for i := 0; i < len(temp); i++ {
			temp[i] = make([]byte, 1500)
		}
		return temp
	}
}
