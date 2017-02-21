package main

import "github.com/perkjelsvik/TTK4145-sanntid/exercises/ex04/src/bcast"

func main() {
	wait := make(chan int)
	bcast.Udp_receive()
	<-wait
}
