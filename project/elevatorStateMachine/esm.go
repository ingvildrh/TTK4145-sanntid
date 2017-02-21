package elevatorStateMachine

import (
	"fmt"

	. "github.com/perkjelsvik/TTK4145-Sanntid/Project/constants"
	hw "github.com/perkjelsvik/TTK4145-Sanntid/Project/hardware"
)

const (
	idle int = iota
	moving
	doorOpen
)

var state int
var floor int
var dir Direction
var queue [NumFloors][NumButtons]int

type Channels struct {
	OrderComplete  chan int
	ElevatorState  chan int
	StateError     chan error
	LocalQueue     chan [NumFloors][NumButtons]int
	NewOrderChan   chan Keypress
	ArrivedAtFloor chan int
	doorTimeout    chan bool
}

func ESM_init(ch Channels, btnsPressed chan Keypress) {
	state = idle
	dir = DirStop
	floor = hw.GetFloorSensorSignal()
	go ESM_loop(ch, btnsPressed)
}

func ESM_loop(ch Channels, btnsPressed chan Keypress) {
	for {
		select {
		case newOrder := <-btnsPressed:
			// Everytime new order, update local orders
			// If idle, execute order
			NewOrder(newOrder)
		case floor = <-ch.ArrivedAtFloor:
			// Everytime we stop at floor, check if order complete
			// if order complete, stop elevator etc.
			ArrivedAtFloor(ch)
			fmt.Println("Arrived at floor")
		case <-ch.doorTimeout:
			// Order complete, send confirmation
			// check if any more orders, go to idle if not
			state = idle
			fmt.Println("Door timeout")
		}
	}
}

// NOTE: Should not be public, but is public to test from main
func NewOrder(newOrder Keypress) {
	fmt.Println("New order")
	queue[newOrder.Floor][newOrder.Btn] = 1
	switch state {
	case idle:
		dir = chooseDirection(newOrder)
		hw.SetMotorDirection(dir)
		if dir == DirStop {
			state = idle
		} else {
			state = moving
		}
	case moving:
	case doorOpen:
		//TODO: Implement door-timer-reset if keypress at current floor
	default:
		fmt.Println("default error")
	}
}

func ArrivedAtFloor(ch Channels) {
	// TODO: send order complete to governor
	// TODO: start doorTimer and doorLight
	// TODO: Move into shouldStop-function instead
	if queue[floor][BtnUp] == 1 {
		queue[floor][BtnUp] = 0
		hw.SetMotorDirection(DirStop)
		dir = DirStop
		// FIXME: State should not be idle
		state = idle
		hw.SetButtonLamp(BtnUp, floor, 0)
		// QUESTION: Should doorTimeout be part of ch? or global var?
		// FIXME: Doesn't activate event in ESM_loop
		//ch.doorTimeout <- true
	} else if queue[floor][BtnDown] == 1 {
		queue[floor][BtnDown] = 0
		hw.SetMotorDirection(DirStop)
		dir = DirStop
		state = idle
		hw.SetButtonLamp(BtnDown, floor, 0)
		//ch.doorTimeout <- true
	} else if queue[floor][BtnInside] == 1 {
		queue[floor][BtnInside] = 0
		hw.SetMotorDirection(DirStop)
		dir = DirStop
		state = idle
		hw.SetButtonLamp(BtnInside, floor, 0)
		//ch.doorTimeout <- true
	}
}

func chooseDirection(key Keypress) Direction {
	if floor > key.Floor {
		return DirDown
	} else if floor < key.Floor {
		return DirUp
	} else {
		return DirStop
	}
}
