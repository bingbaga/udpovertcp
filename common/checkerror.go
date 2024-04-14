package common

import "net"

func IsConnectionClosed(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		// Check if the error is related to a closed connection
		return opErr.Err.Error() == "use of closed network connection"
	}
	return false
}
