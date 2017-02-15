package bcast

import (
	"fmt"
	"net"

	"localip"
)

func Udp_receive() {
	buffer := make([]byte, 1024)
	ServerAddr, _ := net.ResolveUDPAddr("udp", ":30000")
	conn, _ := net.ListenUDP("udp", ServerAddr)
	var localIP, _ = localip.LocalIP()
	fmt.Println("LOCAL IP: " + localIP)
	for {
		fmt.Println("step1")
		n, addr, err := conn.ReadFromUDP(buffer)
		fmt.Println("\tHER: ", err)
		fmt.Println("step2")
		if addr.String() != localIP {
			fmt.Println(string(buffer[0:n]))
		}
	}
	//defer conn.Close()

}

/*func main() {
	udp_receive(":30000")
	ch := make(chan int, 1)
	<-ch
}
*/
