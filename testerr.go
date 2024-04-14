package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("udp", "223.5.5.5:53")
	if err != nil {
		fmt.Println(err.Error())
	}
	time.AfterFunc(time.Second*2, func() {
		conn.Close()
	})
	for {
		time.Sleep(time.Second)
		_, err = conn.Write([]byte("dfasdfasdfsdfsdfsdfsadfsdfsadfasdfs"))
		if err != nil {
			fmt.Printf("isConnectionClosed(err): %v\n", isConnectionClosed(err))
			return
		}
	}

}

func isConnectionClosed(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		// Check if the error is related to a closed connection
		return opErr.Err.Error() == "use of closed network connection"
	}
	return false
}
