package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os/exec"
	"time"
)

func main() {
	fmt.Println("Program B")
	// Create process A
	err := exec.Command("gnome-terminal", "--", "go", "run", "../exercise4programA/main.go").Run()
	if err == nil {
		fmt.Println("Started A")
	}

	conn, err := net.Dial("udp", "localhost:20000")
	if err != nil {
		fmt.Println("Error in dial")
	}
	defer conn.Close()
	for {
		buf := new(bytes.Buffer)
		err = binary.Write(buf, binary.LittleEndian, int32(42))
		if err != nil {
			fmt.Println(err)
		}
		conn.Write(buf.Bytes())
		time.Sleep(1 * time.Second)
	}
}
