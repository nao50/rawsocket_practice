package main

import (
	"fmt"
	"log"
	"net"
	"syscall"
)

func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}

func main() {
	///////////////////////////////////////////////////////////////////////////////////////
	// common
	proto := htons(syscall.ETH_P_ALL)
	buffer := make([]byte, 1500)

	///////////////////////////////////////////////////////////////////////////////////////
	// recv
	recvSockFd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, proto)
	if err != nil {
		log.Fatal("recvSockFd: ", err)
	}
	defer syscall.Close(recvSockFd)

	// syscall.SetsockoptInt(recvSockFd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)

	recvIf, err := net.InterfaceByName("ens4")
	if err != nil {
		log.Fatal("interfacebyname: ", err)
	}

	// var haddr [8]byte
	// copy(haddr[0:7], recvIf.HardwareAddr[0:7])
	recvSll := syscall.SockaddrLinklayer{
		Protocol: proto,
		Ifindex:  recvIf.Index,
		// Halen:    uint8(len(recvIf.HardwareAddr)),
		// Addr:     haddr,
	}
	if err := syscall.Bind(recvSockFd, &recvSll); err != nil {
		log.Fatal("bind: ", err)
	}

	///////////////////////////////////////////////////////////////////////////////////////
	// send
	sendSockFd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, proto)
	if err != nil {
		log.Fatal("sendSockFd: ", err)
	}
	defer syscall.Close(sendSockFd)

	// syscall.SetsockoptInt(sendSockFd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)

	sendIf, err := net.InterfaceByName("ens5")
	if err != nil {
		log.Fatal("interfacebyname: ", err)
	}

	sendSll := syscall.SockaddrLinklayer{
		Protocol: proto,
		Ifindex:  recvIf.Index,
		// Halen:    uint8(len(recvIf.HardwareAddr)),
		// Addr:     haddr,
	}
	if err := syscall.Bind(sendSockFd, &sendSll); err != nil {
		log.Fatal("bind: ", err)
	}

	///////////////////////////////////////////////////////////////////////////////////////
	// main loop
	for {
		n, addr, err := syscall.Recvfrom(recvSockFd, buffer, 0)
		if err != nil {
			log.Fatalln(err)
		}
		// FOR DEBUG
		a := addr.(*syscall.SockaddrInet4)
		fmt.Printf("Found peer %v.%v.%v.%v:%v\n", a.Addr[0], a.Addr[1], a.Addr[2], a.Addr[3], a.Port)

		go func() {
			err = syscall.Sendto(sendSockFd, buffer[:n], 0, &sendSll)
			if err != nil {
				log.Fatalln(err)
			}
		}()

	}

}
