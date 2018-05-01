package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"syscall"
)

/////////////////////////////////////////////////////////////////////////////
type EtherHeader struct {
	Dst  net.HardwareAddr // destination address
	Src  net.HardwareAddr // source address
	Type int              // type
}

// Parse parses b as an Erther header and sotres the result in h.
func (h *EtherHeader) Parse(b []byte) error {
	h.Dst = net.HardwareAddr(b[0:6])
	h.Src = net.HardwareAddr(b[6:12])
	h.Type = int(binary.BigEndian.Uint16(b[12:14]))
	return nil
}

// ParseHeader parses b as an IPv4 header.
func ParseEtherHeader(b []byte) (*EtherHeader, error) {
	h := new(EtherHeader)
	if err := h.Parse(b); err != nil {
		return nil, err
	}
	return h, nil
}

/////////////////////////////////////////////////////////////////////////////
const (
	Version      = 4  // protocol version
	IpHeaderLen  = 20 // header length without extension headers
	maxHeaderLen = 60 // sensible default, revisit if later RFCs define new usage of version and header length fields
)

type IpHeaderFlags int

const (
	MoreFragments IpHeaderFlags = 1 << iota // more fragments flag
	DontFragment                            // don't fragment flag
)

// A Header represents an IPv4 header.
type IpHeader struct {
	Version  int           // protocol version
	Len      int           // header length
	TOS      int           // type-of-service
	TotalLen int           // packet total length
	ID       int           // identification
	Flags    IpHeaderFlags // flags
	FragOff  int           // fragment offset
	TTL      int           // time-to-live
	Protocol int           // next protocol
	Checksum int           // checksum
	Src      net.IP        // source address
	Dst      net.IP        // destination address
	Options  []byte        // options, extension headers
}

// Parse parses b as an IPv4 header and sotres the result in h.
func (h *IpHeader) Parse(b []byte) error {
	if h == nil || len(b) < IpHeaderLen {
		return errors.New("header too short")
	}
	hdrlen := int(b[0]&0x0f) << 2
	if hdrlen > len(b) {
		return errors.New("header too short")
	}
	h.Version = int(b[0] >> 4)
	h.Len = hdrlen
	h.TOS = int(b[1])
	h.ID = int(binary.BigEndian.Uint16(b[4:6]))
	h.TTL = int(b[8])
	h.Protocol = int(b[9])
	h.Checksum = int(binary.BigEndian.Uint16(b[10:12]))
	h.Src = net.IPv4(b[12], b[13], b[14], b[15])
	h.Dst = net.IPv4(b[16], b[17], b[18], b[19])
	h.TotalLen = int(binary.BigEndian.Uint16(b[2:4]))
	h.FragOff = int(binary.BigEndian.Uint16(b[6:8]))
	h.Flags = IpHeaderFlags(h.FragOff&0xe000) >> 13
	h.FragOff = h.FragOff & 0x1fff
	optlen := hdrlen - IpHeaderLen
	if optlen > 0 && len(b) >= hdrlen {
		if cap(h.Options) < optlen {
			h.Options = make([]byte, optlen)
		} else {
			h.Options = h.Options[:optlen]
		}
		copy(h.Options, b[IpHeaderLen:hdrlen])
	}
	return nil
}

// ParseHeader parses b as an IPv4 header.
func ParseIpHeader(b []byte) (*IpHeader, error) {
	h := new(IpHeader)
	if err := h.Parse(b); err != nil {
		return nil, err
	}
	return h, nil
}

/////////////////////////////////////////////////////////////////////////////
func main() {

	// littleendian -> bigendian
	const proto = (syscall.ETH_P_ALL<<8)&0xff00 | syscall.ETH_P_ALL>>8

	// crate sendGTPUSocketFd
	sendGTPUSocketFd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		log.Fatal("socket: ", err)
	}
	defer syscall.Close(sendGTPUSocketFd)

	// get interface struct
	if_send, err := net.InterfaceByName("ens5")
	if err != nil {
		log.Fatal("interfacebyname: ", err)
	}

	sll_send := syscall.SockaddrLinklayer{
		//Protocol: syscall.IPPROTO_UDP,
		Protocol: proto,
		Ifindex:  if_send.Index,
	}
	if err := syscall.Bind(sendGTPUSocketFd, &sll_send); err != nil {
		log.Fatal("bind: ", err)
	}

	/////////////////////////////////////////////////////////////////////////////

	// littleendian -> bigendian
	// const proto = (syscall.ETH_P_ALL<<8)&0xff00 | syscall.ETH_P_ALL>>8
	// crate receiveRawSocketFd
	receiveRawSocketFd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, proto)
	if err != nil {
		log.Fatal("socket: ", err)
	}
	defer syscall.Close(receiveRawSocketFd)

	// get interface struct
	if_recv, err := net.InterfaceByName("ens4")
	if err != nil {
		log.Fatal("interfacebyname: ", err)
	}

	//bind
	var haddr [8]byte
	copy(haddr[0:7], if_recv.HardwareAddr[0:7])
	sll_recv := syscall.SockaddrLinklayer{
		Protocol: proto,
		Ifindex:  if_recv.Index,
		Halen:    uint8(len(if_recv.HardwareAddr)),
		Addr:     haddr,
	}
	if err := syscall.Bind(receiveRawSocketFd, &sll_recv); err != nil {
		log.Fatal("bind: ", err)
	}

	// set promisecas mode
	err = syscall.SetLsfPromisc("ens4", true)
	if err != nil {
		fmt.Println(err)
	}
	// cansel promisecas mode
	err = syscall.SetLsfPromisc("ens4", false)
	if err != nil {
		fmt.Println(err)
	}

	//syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)

	buf := make([]byte, 65536)
	log.Println("Starting raw server...")

	/////////////////////////////////////////////////////////////////////////////

	for {
		n, peer, err := syscall.Recvfrom(receiveRawSocketFd, buf, 0)
		if err != nil {
			log.Fatal("recvfrom ", err)
		}

		sa, _ := peer.(*syscall.SockaddrLinklayer)

		etherheader, err := ParseEtherHeader(buf[0:14])
		if err != nil {
			log.Fatal("ParseEtherHeader error ", err)
		}

		// sa.Pkttype != 4は自ホストから送信されてループバックしてきたパケット（PACKET_OUTGOING）として破棄
		// etherheader.Type != 2054はARP 0x0806 を廃棄
		if sa.Pkttype != 4 && etherheader.Type != 2054 {
			go func() {

				err := syscall.Sendto(sendGTPUSocketFd, buf[:n], 0, &syscall.SockaddrLinklayer{
					//Protocol: syscall.IPPROTO_UDP,
					Protocol: proto,
					Ifindex:  if_send.Index,
				})
				if err != nil {
					fmt.Println(err)
				}

				fmt.Println("==================================")
				fmt.Printf("ALL: %02x \n", buf[:n])
				fmt.Println("==================================")

			}()
		}
	}
}
