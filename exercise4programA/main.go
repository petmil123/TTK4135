package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("Program A")
	var i int32 = -1
	pc, err := net.ListenPacket("udp", ":20000")
	if err != nil {
		fmt.Println("Error in listening")
	}
	defer pc.Close()
	for {
		pc.SetReadDeadline(time.Now().Add(time.Second * 2))
		buf := make([]byte, 4)
		_, _, err := pc.ReadFrom(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, &i)
		fmt.Printf("%d", i)
	}
}
