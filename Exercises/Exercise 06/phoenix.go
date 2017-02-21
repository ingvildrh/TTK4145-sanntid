package phoenix

import (
	"fmt"
	"log"
	"net"
	t "time"
)

var port = 9999
var ipAddr = "127.0.0.1"
var buf = make([]byte, 1024)

func main() {
	//spawnBackup()
	addr, _ := net.ResolveUDPAddr("udp", ipAddr+":"+string(port))
	isPrimary := false
	conn, _ := net.ListenUDP("udp", addr)
	// Error handling?
	log.Println("Noe")

	for !(isPrimary) {
		conn.SetReadDeadline(t.Now().Add(1 * t.Second))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			isPrimary = true
		} else {

		}
	}
	fmt.Println("I'm now primary")
}
