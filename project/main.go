package main

import (
	"time"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorGovernor"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorStateMachine"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/networkCommunication/network/bcast"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/syncElevators"
)

func main() {
	e := ET_Comedi
	esmChans := Channels{
		OrderComplete: make(chan int),
		ElevatorChan:  make(chan Elev),
		//StateError:     make(chan error),
		NewOrderChan:   make(chan Keypress),
		ArrivedAtFloor: make(chan int),
	}
	syncChans := SyncChannels{
		UpdateGovernor: make(chan [NumElevators]Elev),
		UpdateSync:     make(chan Elev),
		IncomingMsg:    make(chan Message),
		OutgoingMsg:    make(chan Message),
		OrderUpdate:    make(chan Keypress),
	}
	btnsPressed := make(chan Keypress)
	syncBtnLights := make(chan [NumFloors][NumButtons]bool)

	HW_init(e, btnsPressed, esmChans.ArrivedAtFloor)

	//TODO: make [NumElevators]Elev it's own type
	ID := 0

	go ESM_loop(esmChans, btnsPressed)
	go GOV_loop(ID, esmChans, btnsPressed, syncChans.UpdateSync, syncChans.UpdateGovernor, syncChans.OrderUpdate, syncBtnLights)
	go GOV_lightsLoop(syncBtnLights)
	go Transmitter(16569, syncChans.OutgoingMsg)
	go Receiver(16569, syncChans.IncomingMsg)
	go SYNC_loop(syncChans, ID)

	elevator := Elev{
		State: 0,
		Dir:   DirStop,
		Floor: 1,
		Queue: [NumFloors][NumButtons]bool{},
	}

	elevator.Queue[2][BtnUp] = true

	var elevList [NumElevators]Elev
	elevList[ID] = elevator
	elevList[ID].Queue[2][BtnUp] = true

	var regOrders [NumFloors]AckMatrix

	regOrders[2].OrderUp = true
	regOrders[2].OrderDown = false
	regOrders[2].DesignatedElevator = 2
	regOrders[2].ImplicitAcks[2] = Acked

	helloMsg := Message{
		Elevator:         elevList,
		RegisteredOrders: regOrders,
	}

	for i := 0; i < 100; i++ {
		syncChans.IncomingMsg <- helloMsg
		time.Sleep(2 * time.Second)
	}
	select {}
}
