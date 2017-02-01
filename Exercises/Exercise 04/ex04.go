package main

import "./src/bcast"

func main() {
	wait := make(chan int)
	bcast.Udp_receive()
	<-wait
}
