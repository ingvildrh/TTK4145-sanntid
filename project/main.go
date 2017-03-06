package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorGovernor"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorStateMachine"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/networkCommunication/network/bcast"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/syncElevators"
)

func main() {
	e := ET_Simulation
	esmChans := Channels{
		OrderComplete: make(chan int, (NumFloors*NumButtons)-2),
		ElevatorChan:  make(chan Elev, 10),
		//StateError:     make(chan error),
		NewOrderChan:   make(chan Keypress, (NumFloors*NumButtons)-2),
		ArrivedAtFloor: make(chan int),
	}
	syncChans := SyncChannels{
		UpdateGovernor: make(chan [NumElevators]Elev),
		UpdateSync:     make(chan Elev),
		IncomingMsg:    make(chan Message, 10),
		OutgoingMsg:    make(chan Message, 10),
		OrderUpdate:    make(chan Keypress),
	}
	btnsPressed := make(chan Keypress)
	syncBtnLights := make(chan [NumFloors][NumButtons]bool, 10)

	HW_init(e, btnsPressed, esmChans.ArrivedAtFloor)

	//TODO: make [NumElevators]Elev it's own type
	ID := 1
	go ESM_loop(esmChans, btnsPressed)
	go GOV_loop(ID, esmChans, btnsPressed, syncChans.UpdateSync, syncChans.UpdateGovernor, syncChans.OrderUpdate, syncBtnLights)
	go GOV_lightsLoop(syncBtnLights)
	go Transmitter(9996, syncChans.OutgoingMsg)
	go Receiver(9997, syncChans.IncomingMsg)
	go SYNC_loop(syncChans, ID)

	elevator := Elev{
		State: 0,
		Dir:   DirStop,
		Floor: 1,
		Queue: [NumFloors][NumButtons]bool{},
	}

	var elevList [NumElevators]Elev
	elevList[1] = elevator

	var regOrders [NumFloors][NumButtons - 1]AckList

	regOrders[2][1].DesignatedElevator = 1
	regOrders[2][1].ImplicitAcks[1] = Acked
	regOrders[0][0].DesignatedElevator = 0
	regOrders[0][0].ImplicitAcks[1] = Acked

	helloMsg := Message{
		Elevator:         elevList,
		RegisteredOrders: regOrders,
	}

	go killSwitch()

	for {
		elevator.Queue[2][BtnDown] = true
		time.Sleep(5000 * time.Millisecond)
		syncChans.OutgoingMsg <- helloMsg
		elevator.Queue[2][BtnDown] = false
		time.Sleep(5000 * time.Millisecond)
		elevator.Queue[0][BtnUp] = true
		syncChans.OutgoingMsg <- helloMsg
		elevator.Queue[0][BtnUp] = false
	}

	//select {}
}

func killSwitch() {
	// safeKill turns the motor off if the program is killed with CTRL+C.
	var c = make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
	SetMotorDirection(DirStop)
	fmt.Println("\x1b[31;1m", "User terminated program.", "\x1b[0m")
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			SetStopLamp(1)
		} else {
			SetStopLamp(0)
		}
		time.Sleep(500 * time.Millisecond)
	}
	os.Exit(1)
}
