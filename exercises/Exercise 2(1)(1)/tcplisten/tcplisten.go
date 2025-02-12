package main

import (
	"fmt"
	"net"
)

// 10.100.23.204
func main() {
	own_ip := "10.100.23.28:20018"
	conn, err := net.Dial("tcp", "10.100.23.204:33546") //establish connection with TCP server
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	conn.Write([]byte("Connect to: " + own_ip + "\000")) //ask the tcp server to connect with our IP addr

	ln, err := net.Listen("tcp", "10.100.23.28:20018") //
	if err != nil {
		fmt.Println(err)
	}
	conn2, err := ln.Accept()
	if err != nil {
		fmt.Println("Error in accept")
	}
	for {
		buf := make([]byte, 1024)
		buf2 := make([]byte, 1024)

		conn.Read(buf)
		fmt.Printf("%s", buf)

		conn2.Read(buf2)
		fmt.Printf("20018: %s", buf2)

	}
}
