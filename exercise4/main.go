package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os/exec"
	"time"
)

func mainLoop(i int32) {
	conn, err := net.Dial("udp", "localhost:20000")
	if err != nil {
		fmt.Println("Error in dial")
	}
	defer conn.Close()
	for {
		// Send state on UDP
		buf := new(bytes.Buffer)
		err = binary.Write(buf, binary.LittleEndian, int32(i))
		conn.Write(buf.Bytes())
		//Look for response
		//if no response, create backup

		time.Sleep(1 * time.Second)
		fmt.Println(i)
		i += 1
	}
}

func backupLoop(i int32) {
	// Read state from UDP, if no answer within some seconds, take over
	pc, err := net.ListenPacket("udp", ":20000")
	if err != nil {
		fmt.Println("Error in listening")
	}
	defer pc.Close()
	for {
		pc.SetReadDeadline(time.Now().Add(time.Second * 3))
		buf := make([]byte, 4)
		_, _, err := pc.ReadFrom(buf)
		if err != nil {
			fmt.Println(err)
			pc.Close()
			err := exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()
			if err == nil {
				fmt.Println("Started A")
			}
			break
		}
		err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, &i)
	}
	mainLoop(i)
	//try read from UDP
	//if fail:
	//createNewBackup
	//mainLoop(i)

}

func main() {
	backupLoop(0)
}
