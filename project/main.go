package main

import (
	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorGovernor"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorStateMachine"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/networkCommunication/network/bcast"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/syncElevators"
)

func main() {
	e := ET_Comedi
	ch := Channels{
		OrderComplete:  make(chan int),
		ElevatorChan:   make(chan Elev),
		StateError:     make(chan error),
		NewOrderChan:   make(chan Keypress),
		ArrivedAtFloor: make(chan int),
	}
	btnsPressed := make(chan Keypress)
	NetworkUpdate := make(chan int)
	syncBtnLights := make(chan bool)
	incomingMsg := make(chan Msg)
	outoingMsg := make(chan Msg)

	HW_init(e, btnsPressed, ch.ArrivedAtFloor)
	//TODO: NetWorkUpdate channel to governor
	ID := 0

	go ESM_loop(ch, btnsPressed)
	go GOV_loop(ID, ch, btnsPressed, NetworkUpdate, syncBtnLights)
	go GOV_lightsLoop(syncBtnLights)
	go Transmitter(16569, outoingMsg)
	go Receiver(16569, incomingMsg)
	go SYNC_loop(NetworkUpdate, incomingMsg, ID)
	go SYNC_periodicStateSpammer(outoingMsg)
	select {}
}
