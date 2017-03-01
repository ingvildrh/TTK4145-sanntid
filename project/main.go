package main

import (
	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	esm "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorStateMachine"
	hw "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
)

func main() {
	e := hw.ET_Comedi
	ch := esm.Channels{
		OrderComplete:  make(chan int),
		ElevatorState:  make(chan int),
		StateError:     make(chan error),
		NewOrderChan:   make(chan Keypress),
		ArrivedAtFloor: make(chan int),
	}
	btnsPressed := make(chan Keypress)
	hw.HW_init(e, btnsPressed, ch.ArrivedAtFloor)
	go esm.ESM_loop(ch, btnsPressed)
	//TODO: NetWorkUpdate channel to governor
	//gov.GOV_init()
	select {}
}
