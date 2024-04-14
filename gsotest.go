package main

import (
	"encoding/binary"
	"fmt"
	"gso/udp"
	"sync/atomic"
	"time"

	"log"
	"net"

	"syscall"

	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"
)

const (
	UDP_SEGMENT = 103
)

func getCmsg(size int) []byte {
	b := make([]byte, 8+4+4+2)
	binary.LittleEndian.PutUint64(b[:8], 18)
	binary.LittleEndian.PutUint32(b[8:12], uint32(syscall.IPPROTO_UDP))
	binary.LittleEndian.PutUint32(b[12:16], uint32(UDP_SEGMENT))
	binary.LittleEndian.PutUint16(b[16:], uint16(size))
	return b
}
func supportsUDPOffload1(conn *net.UDPConn) (txOffload, rxOffload bool) {
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
func batchUdp() {
	conn := udp.NewUdpListen(&net.UDPAddr{
		IP:   net.IPv4(192, 168, 1, 23),
		Port: 22229,
	}, 1400, 35)
	var count int64 = 0

	msgs := make([]ipv4.Message, 35)
	go func() {
		for {
			time.Sleep(time.Second)
			n := atomic.LoadInt64(&count)
			atomic.SwapInt64(&count, 0)
			if n > (1 << 20) {
				fmt.Println(n/1024/1024, "MBps")
			} else {
				fmt.Println(n/1024, "KBps")
			}

		}

	}()
	for v := range msgs {
		msgs[v].Buffers = make([][]byte, 1)
		for y := range msgs[v].Buffers {
			msgs[v].Buffers[y] = make([]byte, 1500)
		}
	}
	// bb := bytes.Buffer{}
	for {
		n, err := conn.Conn.ReadBatch(msgs, 0)
		if err != nil {
			fmt.Println(err.Error())
		}

		for i := 0; i < n; i++ {
			length := msgs[i].N
			if length == 0 {
				continue
			}
			atomic.AddInt64(&count, int64(length))
			// bb.Write(msgs[i].Buffers[0][:length])
			// bb.Reset()
		}
	}

}
func NormalUdp() {

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(192, 168, 1, 23),
		Port: 22229,
	})
	fmt.Println(supportsUDPOffload1(conn))
	if err != nil {
		log.Fatal(err.Error())
	}
	// bb := bytes.Buffer{}
	b := make([]byte, 1500)
	var count int64 = 0
	go func() {
		for {
			time.Sleep(time.Second)
			n := atomic.LoadInt64(&count)
			atomic.SwapInt64(&count, 0)
			if n > (1 << 20) {
				fmt.Println(n/1024/1024, "MBps")
			} else {
				fmt.Println(n/1024, "KBps")
			}

		}

	}()
	for {
		n, err := conn.Read(b)
		if err != nil {
			fmt.Println(err.Error())
		}
		if n > 18 {
			fmt.Println(n)
			return
		}

		atomic.AddInt64(&count, int64(n))
		// bb.Write(b[:n])

		// bb.Reset()

	}
}

func main() {
	// batchUdp()
	NormalUdp()

	// conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
	// 	IP:   net.IPv4(127, 0, 0, 1),
	// 	Port: 22233,
	// })
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	// conn.SetWriteBuffer(1024 * 1024 * 7)
	// conn.SetReadBuffer(1024 * 1024 * 7)
	// packetconn := ipv4.NewPacketConn(conn)

	// b := make([]byte, 18*35)
	// msgs := make([]ipv4.Message, 1)
	// for v := range msgs {
	// 	msgs[v].Buffers = make([][]byte, 1)
	// 	for y := range msgs[v].Buffers {
	// 		msgs[v].Buffers[y] = b
	// 	}
	// 	msgs[v].OOB = getCmsg(18)
	// 	//
	// }
	// now := time.Now()
	// for i := 0; i < 35; i++ {
	// 	// conn.Write(b)
	// 	_, err := packetconn.WriteBatch(msgs, 0)
	// 	if err != nil {
	// 		fmt.Println(err.Error())
	// 	}
	// }
	// fmt.Println(time.Since(now))
	// time.Sleep(time.Second)

}
