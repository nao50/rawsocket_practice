package main

import (
	"fmt"
	"log"
	"syscall"
)

func htons(i int) int {
	return (i<<8)&0xff00 | i>>8
}

func main() {
	///////////////////////////////////////////////////////////////////////////////////////
	// common
	// const proto = (syscall.IPPROTO_UDP<<8)&0xff00 | syscall.IPPROTO_UDP>>8
	// const proto = (syscall.ETH_P_ALL<<8)&0xff00 | syscall.ETH_P_ALL>>8
	// proto := htons(syscall.ETH_P_ALL)
	// buffer := make([]byte, 1500)

	///////////////////////////////////////////////////////////////////////////////////////
	// send
	// sendSockFd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, proto)
	// sendSockFd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, 0)
	sendSockFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		fmt.Println("aaaaaaaaa")
		log.Fatal("sendSockFd: ", err)
	}
	defer syscall.Close(sendSockFd)

	// syscall.SetsockoptInt(sendSockFd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)

	// sendIf, err := net.InterfaceByName("ens5")
	// if err != nil {
	// 	log.Fatal("interfacebyname: ", err)
	// }

	// sendSll := syscall.SockaddrLinklayer{
	// 	// Protocol: proto,
	// 	Protocol: syscall.IPPROTO_RAW,
	// 	Ifindex:  sendIf.Index,
	// 	// Halen:    uint8(len(recvIf.HardwareAddr)),
	// 	// Addr:     haddr,
	// }
	// if err := syscall.Bind(sendSockFd, &sendSll); err != nil {
	// 	log.Fatal("bind: ", err)
	// }

	addr := &syscall.SockaddrInet4{Port: 2152, Addr: [4]byte{172, 20, 100, 2}}

	///////////////////////////////////////////////////////////////////////////////////////
	// main loop
	buff := []byte("test012345")
	// err = syscall.Sendto(sendSockFd, buff, 0, &sendSll)
	err = syscall.Sendto(sendSockFd, buff, 0, addr)
	if err != nil {
		log.Fatalln(err)
	}

	// fmt.Println("Starting raw server...")
	// for {
	// 	go func() {
	// 		buff := []byte("test012345")
	// 		// err = syscall.Sendto(sendSockFd, buffer[:n], 0, &sendSll)
	// 		err := syscall.Sendto(sendSockFd, buff, 0, addr)
	// 		if err != nil {
	// 			log.Fatalln(err)
	// 		}
	// 	}()
	// }

}
