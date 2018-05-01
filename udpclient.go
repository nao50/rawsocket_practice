package main

import (
	"net"
)

func main() {
	req := []byte("test012345")
	conn, err := net.Dial("udp4", "localhost:2152")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.Write(req)
	if err != nil {
		panic(err)
	}

	// buffer := make([]byte, 1500)
	// length, err := conn.Read(buffer)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(buffer[:length]))

}
