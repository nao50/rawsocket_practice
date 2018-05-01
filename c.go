package main

import (
	"log"
	"net"
	"syscall"

	"golang.org/x/net/ipv4"
)

func main() {
	h := ipv4.Header{
		Version:  4,
		Len:      20,
		TotalLen: 20 + 10, // 20 bytes for IP, 10 for ICMP
		TTL:      64,
		Protocol: 1, // ICMP
		Dst:      net.IPv4(172, 20, 100, 2),
		// ID, Src and Checksum will be set for us by the kernel
	}
	out, err := h.Marshal()

	icmp := []byte{
		8, // type: echo request
		0, // code: not used by echo request
		0, // checksum (16 bit), we fill in below
		0,
		0, // identifier (16 bit). zero allowed.
		0,
		0, // sequence number (16 bit). zero allowed.
		0,
		0xC0, // Optional data. ping puts time packet sent here
		0xDE,
	}

	buff := append(out, icmp...)
	// buff := append(icmp..., out)

	fd, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	defer syscall.Close(fd)

	err = syscall.SetsockoptString(fd, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, "ens5")
	if err != nil {
		log.Fatal("Sendto:", err)
	}

	addr := syscall.SockaddrInet4{
		Port: 0,
		Addr: [4]byte{127, 0, 0, 2},
	}

	// p := pkt()
	err = syscall.Sendto(fd, buff, 0, &addr)
	// err = syscall.Sendto(fd, buff, 0, &sendSll)
	if err != nil {
		log.Fatal("Sendto:", err)
	}
}

// func pkt() []byte {
// 	h := ipv4.Header{
// 		Version:  4,
// 		Len:      20,
// 		TotalLen: 20 + 10, // 20 bytes for IP, 10 for ICMP
// 		TTL:      64,
// 		Protocol: 1, // ICMP
// 		Dst:      net.IPv4(127, 0, 0, 1),
// 		// ID, Src and Checksum will be set for us by the kernel
// 	}

// 	icmp := []byte{
// 		8, // type: echo request
// 		0, // code: not used by echo request
// 		0, // checksum (16 bit), we fill in below
// 		0,
// 		0, // identifier (16 bit). zero allowed.
// 		0,
// 		0, // sequence number (16 bit). zero allowed.
// 		0,
// 		0xC0, // Optional data. ping puts time packet sent here
// 		0xDE,
// 	}
// 	cs := csum(icmp)
// 	icmp[2] = byte(cs)
// 	icmp[3] = byte(cs >> 8)

// 	out, err := h.Marshal()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return append(out, icmp...)
// }

// func csum(b []byte) uint16 {
// 	var s uint32
// 	for i := 0; i < len(b); i += 2 {
// 		s += uint32(b[i+1])<<8 | uint32(b[i])
// 	}
// 	// add back the carry
// 	s = s>>16 + s&0xffff
// 	s = s + s>>16
// 	return uint16(^s)
// }
