package main

import (
	"fmt"
	"net"
)

func main() {
	buffer := make([]byte, 1024)
	ServerAddr, _ := net.ResolveUDPAddr("udp", ":30000")
	conn, _ := net.ListenUDP("udp", ServerAddr)
	defer conn.Close()
	var localIP = "129.241.187.159"
	for {
		n, addr, _ := conn.ReadFromUDP(buffer)

		if addr.String() != localIP {
			fmt.Println(string(buffer[0:n]))
		}
	}
}
