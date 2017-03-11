package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/perkjelsvik/TTK4145-sanntid/exercises/ex04/src/localip"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorGovernor"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorStateMachine"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
	"github.com/perkjelsvik/TTK4145-sanntid/project/networkCommunication/network/bcast"
	"github.com/perkjelsvik/TTK4145-sanntid/project/networkCommunication/network/peers"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/syncElevators"
)

func main() {
	// TODO: Change static arrays to slices (Elev, ElevList etc.)
	// TODO: Change to maps for elevList etc. as well
	// TODO: Handling lost and new peers
	// TODO: Handle crashing
	// TODO: Handle losing network (complete internal orders)
	// TODO: Actually light the lights on all the elevators
	// TODO: General cleanup (rename constants, go through code/declarations etc.)
	elevType := ""
	id := ""
	e := ET_Comedi
	ID := 0
	simPort := ""
	flag.StringVar(&elevType, "run", "", "run type")
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.IntVar(&ID, "ID", 0, "ID of this elevator")
	flag.StringVar(&simPort, "simPort", "9998", "simulation server port")
	flag.Parse()
	if ID != 0 {
		fmt.Println("ID: ", ID)
	}
	if elevType == "sim" {
		e = ET_Simulation
		fmt.Println("Running in simulation mode!")
	}

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	fmt.Println()

	// TODO: Define channels as input/output/bidirectional instead of all bidirectional
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
		PeerUpdate:     make(chan peers.PeerUpdate),
		PeerTxEnable:   make(chan bool),
	}
	btnsPressed := make(chan Keypress)
	syncBtnLights := make(chan [NumElevators]Elev) //[NumFloors][NumButtons]bool)

	HW_init(e, btnsPressed, esmChans.ArrivedAtFloor, simPort)
	//TODO: make [NumElevators]Elev it's own type
	//IDEA: make peer channels and thread
	//TODO: ID should be the id from above, and then simply use map
	//QUESTION: Should we have inits as functions and then loops as gothreads?
	go ESM_loop(esmChans, btnsPressed)
	go GOV_loop(ID, esmChans, btnsPressed, syncChans.UpdateSync, syncChans.UpdateGovernor, syncChans.OrderUpdate, syncBtnLights)
	go GOV_lightsLoop(syncBtnLights, ID)
	// added syncBtnLights
	go SYNC_loop(syncChans, ID) //, syncBtnLights)

	go peers.Transmitter(42039, id, syncChans.PeerTxEnable)
	go peers.Receiver(42039, syncChans.PeerUpdate)
	go bcast.Transmitter(42034, syncChans.OutgoingMsg)
	go bcast.Receiver(42034, syncChans.IncomingMsg)
	/*
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
	*/
	go killSwitch()
	/*
		for {gffrtgf
			elevator.Queue[2][BtnDown] = true
			time.Sleep(5000 * time.Millisecond)
			syncChans.OutgoingMsg <- helloMsg
			elevator.Queue[2][BtnDown] = false
			time.Sleep(5000 * time.Millisecond)
			elevator.Queue[0][BtnUp] = true
			syncChans.OutgoingMsg <- helloMsg
			elevator.Queue[0][BtnUp] = false
		}
	*/
	select {}
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
	//Restart.Run()
	os.Exit(1)
}
