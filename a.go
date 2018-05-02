package main

import (
	"fmt"
	"log"
	"net"
	"syscall"
)

func htons(i int) int {
	return (i<<8)&0xff00 | i>>8
}

func main() {
	///////////////////////////////////////////////////////////////////////////////////////
	// common
	// const proto = (syscall.ETH_P_ALL<<8)&0xff00 | syscall.ETH_P_ALL>>8
	const proto = (syscall.ETH_P_IP<<8)&0xff00 | syscall.ETH_P_IP>>8
	// const proto2 = (syscall.IPPROTO_RAW<<8)&0xff00 | syscall.IPPROTO_RAW>>8
	// proto := htons(syscall.ETH_P_ALL)
	buffer := make([]byte, 2500)

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
	fmt.Println("ens4: ", recvIf)

	// var recvIfHaddr [8]byte
	// copy(recvIfHaddr[0:7], recvIf.HardwareAddr[0:7])
	recvSll := syscall.SockaddrLinklayer{
		Protocol: proto,
		Ifindex:  recvIf.Index,
		// Halen:    uint8(len(recvIf.HardwareAddr)),
		// Addr:     recvIfHaddr,
	}
	fmt.Println("recvSll: ", recvSll)
	if err := syscall.Bind(recvSockFd, &recvSll); err != nil {
		log.Fatal("bind: ", err)
	}

	///////////////////////////////////////////////////////////////////////////////////////
	// send
	// sendSockFd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, proto)
	// sendSockFd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	sendSockFd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	// sendSockFd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, proto)
	// sendSockFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		log.Fatal("sendSockFd: ", err)
	}
	defer syscall.Close(sendSockFd)

	// syscall.SetsockoptInt(sendSockFd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)

	sendIf, err := net.InterfaceByName("ens5")
	if err != nil {
		log.Fatal("interfacebyname: ", err)
	}
	fmt.Println("ens5: ", sendIf)

	// var sendIfHaddr [8]byte
	// copy(sendIfHaddr[0:7], sendIf.HardwareAddr[0:7])
	sendSll := syscall.SockaddrLinklayer{
		Protocol: proto,
		Ifindex:  sendIf.Index,
		// Halen:    uint8(len(sendIf.HardwareAddr)),
		// Addr:     sendIfHaddr,
	}
	fmt.Println("sendSll: ", sendSll)
	if err := syscall.Bind(sendSockFd, &sendSll); err != nil {
		log.Fatal("bind: ", err)
	}

	// sendSll := syscall.SockaddrInet4{
	// 	Addr: [4]byte{172, 20, 100, 2},
	// }

	///////////////////////////////////////////////////////////////////////////////////////
	// main loop
	fmt.Println("Starting raw server...")
	for {
		n, addr, err := syscall.Recvfrom(recvSockFd, buffer, 0)
		if err != nil {
			log.Fatalln(err)
		}
		// FOR DEBUG
		// a := addr.(*syscall.SockaddrInet4)
		// fmt.Printf("Found peer %v.%v.%v.%v:%v\n", a.Addr[0], a.Addr[1], a.Addr[2], a.Addr[3], a.Port)
		sa, _ := addr.(*syscall.SockaddrLinklayer)
		fmt.Printf("Recv SockaddrLinklayer: %+v\n", sa)

		fmt.Printf("Recv Buffer: %v\n", buffer[:n])

		go func() {
			err = syscall.Sendto(sendSockFd, buffer[:n], 0, &sendSll)
			// err := syscall.Sendto(sendSockFd, buffer[:n], 0, &syscall.SockaddrLinklayer{
			// 	Protocol: proto,
			// 	Ifindex:  sendIf.Index,
			// })
			if err != nil {
				log.Fatalln(err)
			}
		}()

	}

}
