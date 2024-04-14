package udp

import (
	"bytes"
	"encoding/binary"
	"net"


	"github.com/cespare/xxhash"
)

type Udpdata struct {
	Length  uint16
	DstPort uint16
	DstIP   net.IP
	Data    []byte
}

func (u *Udpdata) ID() uint64 {
	var b []byte
	b = binary.LittleEndian.AppendUint16(b, u.DstPort)
	b = append(b, u.DstIP.To4()...)
	return xxhash.Sum64(b)
}

func (u *Udpdata) Marshal() []byte {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.LittleEndian, u.Length)
	binary.Write(&buf, binary.LittleEndian, u.DstPort)
	buf.Write(u.DstIP.To4())
	buf.Write(u.Data)
	return buf.Bytes()
}
func (u *Udpdata) Len() int {
	return 2 + 2 + 4 + len(u.Data)
}
func NewUdpData(ip net.Addr, data []byte) (*Udpdata, error) {
	newIp, err := net.ResolveUDPAddr(ip.Network(), ip.String())
	if err != nil {
		return nil, err
	}
	
	resp := Udpdata{
		DstIP:   newIp.IP,
		DstPort: uint16(newIp.Port),
		Data:    data,
	}
	resp.Length = 2 + 2 + 4 + uint16(len(data))
	return &resp, nil
}
func (u *Udpdata) FastUnmarshal(b []byte) {
	u.Length = binary.LittleEndian.Uint16(b[:2])
	// u.DstPort = binary.LittleEndian.Uint16(b[2:4])
	// u.DstIP = net.IP(b[4:8])
	u.Data = b[8:u.Length]
}
func (u *Udpdata) Unmarshal(b []byte) {
	u.Length = binary.LittleEndian.Uint16(b[:2])
	u.DstPort = binary.LittleEndian.Uint16(b[2:4])
	u.DstIP = net.IP(b[4:8])
	u.Data = b[8:u.Length]
}
