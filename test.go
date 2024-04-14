package main

import (
	"github.com/lbsystem/udpovertcp/udp"
	"io"
	"log"
	"net"
)

func main() {
	udpconn := udp.NewUdpListen(
		&net.UDPAddr{
			IP:   net.IPv4(192, 168, 1, 23),
			Port: 22226,
		},
		1400,
		35,
	)
	// c := &tls.Config{
	// 	InsecureSkipVerify: true,
	// }
	conn, err := net.Dial("tcp", "[2406:da14:1b03:4500:3956:973a:6ab6:3103]:33366")
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
