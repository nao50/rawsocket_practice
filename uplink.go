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
	const proto = (syscall.ETH_P_IP<<8)&0xff00 | syscall.ETH_P_IP>>8
	// const proto = (syscall.ETH_P_ALL<<8)&0xff00 | syscall.ETH_P_ALL>>8
	// proto := htons(syscall.ETH_P_ALL)
	buffer := make([]byte, 1550)

	///////////////////////////////////////////////////////////////////////////////////////
	// S5:UPLINK:Recv:GTPv1Decap
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 2123,
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalln(err)
	}

	///////////////////////////////////////////////////////////////////////////////////////
	// SGi:UPLINK:Send:RawSocket  socket(AF_INET), SOCK_RAW, IPPROTO_RAW
	// sendSockFd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	sendSockFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		log.Fatal("sendSockFd: ", err)
	}
	defer syscall.Close(sendSockFd)

	sendIf, err := net.InterfaceByName("ens5")
	if err != nil {
		log.Fatal("interfacebyname: ", err)
	}
	var sendIfHaddr [8]byte
	copy(sendIfHaddr[0:7], sendIf.HardwareAddr[0:7])
	sendSll := syscall.SockaddrLinklayer{
		Protocol: proto,
		Ifindex:  sendIf.Index,
		Halen:    uint8(len(sendIf.HardwareAddr)),
		Addr:     sendIfHaddr,
	}

	if err := syscall.Bind(sendSockFd, &sendSll); err != nil {
		log.Fatal("bind: ", err)
	}

	///////////////////////////////////////////////////////////////////////////////////////
	// main loop
	fmt.Println("Starting raw server...")
	for {
		// n, addr, err := syscall.Recvfrom(recvSockFd, buffer, 0)
		n, _, err := udpConn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatalln(err)
		}

		// sa, _ := addr.(*syscall.SockaddrLinklayer)
		// fmt.Printf("SockaddrLinklayer: %+v\n", sa)
		fmt.Printf("Buffer: %v\n", buffer[:n])

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
