// package main

// import (
// 	"fmt"
// 	"net"
// 	"time"
// )

// func main() {
// 	network := "10.100.23.204:20068"
// 	message := []byte("Hello from bench 18!")
// 	pc, err := net.Dial("udp", network)
// 	if err != nil {
// 		fmt.Println("Error in dial!")
// 	}

// 	addr = net.Addr(network)
// 	go serve(pc,addr)
// }

// func serve(pc net.PacketConn, addr net.Addr, buf []byte) {
// 	// buf[2] |= 0x80
// 	pc.WriteTo(buf, addr)
// 	time.Sleep(1 * time.Second)

// }
