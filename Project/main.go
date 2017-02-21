package main

import (
	. "github.com/perkjelsvik/TTK4145-Sanntid/Project/constants"
	esm "github.com/perkjelsvik/TTK4145-Sanntid/Project/elevatorStateMachine"
	. "github.com/perkjelsvik/TTK4145-Sanntid/Project/hardware"
)

func main() {
	e := ET_Comedi
	ch := esm.Channels{
		OrderComplete:  make(chan int),
		ElevatorState:  make(chan int),
		StateError:     make(chan error),
		LocalQueue:     make(chan [NumFloors][NumButtons]int),
		ArrivedAtFloor: make(chan int),
	}
	btnsPressed := make(chan Keypress)
	// QUESTION: Should all inits be Goroutines? How to best structure MAKES and GO's
	HW_init(e, btnsPressed, ch.ArrivedAtFloor)
	esm.ESM_init(ch, btnsPressed)
	//newOrder := Keypress{
	//	Floor: 3,
	//	Btn:   1,
	//}
	//esm.NewOrder(newOrder)
	//	for {
	//		if GetFloorSensorSignal() == 3 {
	//			SetMotorDirection(DirStop)
	//		}
	//	}
	run := make(chan bool)
	<-run
}
