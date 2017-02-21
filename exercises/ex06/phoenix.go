package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	t "time"
)

var counter = 0
var port = 9999
var ipAddr = "127.0.0.1"
var buf = make([]byte, 1024)

func spawnBackup() {
	(exec.Command("gnome-terminal", "-x", "sh", "-c", "go run ~/go/src/github.com/perkjelsvik/TTK4145-sanntid/exercises/ex06/phoenix.go")).Run()

	fmt.Println("New backup up and running")

}

func main() {
	//spawnBackup()
	addr, _ := net.ResolveUDPAddr("udp", ipAddr+":"+string(port))
	isPrimary := false
	conn, _ := net.ListenUDP("udp", addr)
	// Error handling?
	log.Println("Noe")

	for !(isPrimary) {
		conn.SetReadDeadline(t.Now().Add(1 * t.Second))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			isPrimary = true
		} else {
			counter = int(buf[0:n])
		}
	}
	conn.Close()

	spawnBackup()
	fmt.Println("I'm now primary")
	bcastConn, _ := net.DialUDP("udp", nil, addr)

	for {
		fmt.Println(counter)
		counter++
		bcastConn.Write([]byte(buf))
		t.Sleep(1000 * t.Millisecond)
	}
}
