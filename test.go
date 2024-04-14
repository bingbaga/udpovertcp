package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/lbsystem/udpovertcp/common"
	"github.com/lbsystem/udpovertcp/udp"
	"golang.org/x/net/ipv4"
)

func client(localIP, remoteIP string) {
	lIP, err := net.ResolveUDPAddr("udp", localIP)
	if err != nil {
		log.Fatal("localIP is wrong", err.Error())
	}
	udpconn := udp.NewUdpListen(
		lIP,
		1400,
		35,
	)
	// c := &tls.Config{
	// 	InsecureSkipVerify: true,
	// }
	conn, err := net.Dial("tcp", remoteIP)
	if err != nil {
		log.Fatal(err.Error())
	}
	overConn, err := udp.NewTcp(conn, false)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer overConn.Close()

	go func() {
		for {
			reciveData, err := overConn.ReadtoStruct()

			if err != nil {
				// log.Println("tcp decode", err.Error())
				if err == io.EOF {
					break
				}
			}

			if reciveData == nil {
				continue
			}
			udpconn.ClientNormalSend(reciveData)
		}
	}()

	for {
		datalist, err := udpconn.ClientRendToStruct()
		if err != nil {
			log.Fatal(err.Error())
		}

		_, err = overConn.StructToWrite(datalist)
		if err != nil {
			log.Fatal(err.Error())
		}

	}
}

func handleRead(conn *udp.Tcp) {
	fmt.Println("handle start")
	go conn.SendUseChanList()
	defer conn.Close()

	for {

		dataList, err := conn.ReadtoStruct()
		if err != nil {
			fmt.Println("tcp is close", err.Error())
			for k, v := range udp.IdForchanMap {
				v.Close()
				delete(udp.IdForchanMap, k)
			}
			return
		}
		for _, v := range dataList {
			chanforudp, ok := udp.IdForchanMap[v.ID()]
			if ok {
				chanforudp.DataChan <- v
			} else {
				newConn := udp.NewUdpConn(&net.UDPAddr{
					IP:   v.DstIP,
					Port: int(v.DstPort),
				}, 1400, 35)
				go newConn.ServerNormalSend(time.Second * 180)

				newConn.DataChan <- v
				fmt.Println("created new udp conn")
				udp.IdForchanMap[v.ID()] = newConn
				go func() {
					for {
						readMsgs := udp.MessagePool.Get().([]ipv4.Message)
						data := udp.DataListPool.Get().([][]byte)[:0]
						newdata, err := newConn.ServerRendToList(readMsgs, data)
						if err != nil {

							fmt.Printf("common.IsConnectionClosed(err): %v\n", common.IsConnectionClosed(err))
							fmt.Println(err.Error())
							return
						}
						conn.SendListChan <- newdata

					}
				}()
			}
		}

	}

}
func server(localIP, remoteIP string) {

	l, err := net.Listen("tcp", localIP)
	if err != nil {
		log.Fatal(err.Error())
	}
	udp.Dst, err = net.ResolveUDPAddr("udp", remoteIP)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(udp.Dst.String())
	// go func ()  {
	// 	f, err := os.Create("/goProject/udpovertcp/cpu.pprof")
	// 	if err!=nil{
	// 		os.Exit(0)
	// 	}
	// 	pprof.StartCPUProfile(f)
	// 	mysignal:=make(chan os.Signal)
	// 	signal.Notify(mysignal,os.Interrupt,syscall.SIGTERM)
	// 	<-mysignal
	// 	pprof.StopCPUProfile()
	// 	//go tool pprof cpu.pprof
	// 	//(pprof) top 20
	// 	os.Exit(1)
	// }()

	for {
		conn, err := l.Accept()
		fmt.Println(conn.RemoteAddr())
		if err != nil {
			fmt.Println(err.Error())
		}
		overTcp, err := udp.NewTcp(conn, true)
		if err != nil {
			fmt.Println(err.Error())
		}
		go handleRead(overTcp)
	}
}

func main() {
	localIP := flag.String("local", "0.0.0.0:33366", "listen an ip and a port")
	remoteIP := flag.String("remote", "", "It should be a server IP in client mode or a forward target IP in server mode.")
	mode := flag.String("mode", "", "client or server")
	flag.Parse()

	switch *mode {
	case "client":
		log.Println("start client")
		client(*localIP, *remoteIP)
	case "server":
		log.Println("start server")
		server(*localIP, *remoteIP)
	default:
		log.Fatal("mode is wrong")

	}
}
