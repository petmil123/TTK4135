package main

import (
	"fmt"
	"net"
)

func main() {
	// Write
	conn, err := net.Dial("udp", "255.255.255.255:20018")
	if err != nil {
		fmt.Println("error in dial")
	}
	defer conn.Close()
	conn.Write([]byte("Hello from group 18!"))

	// Read
	pc, err := net.ListenPacket("udp", ":20018")
	if err != nil {
		fmt.Println("Error in listen!")
	}
	defer pc.Close()
	for {

		buf := make([]byte, 1024)

		n, _, err := pc.ReadFrom(buf)
		if err != nil {
			fmt.Println("Error in for loop")
		}
		fmt.Printf("%s", buf[:n])
	}

}

// func serve(pc net.PacketConn, addr net.Addr, buf []byte) {
// 	// buf[2] |= 0x80
// 	pc.WriteTo(buf, addr)
// 	time.Sleep(1 * time.Second)

// }
