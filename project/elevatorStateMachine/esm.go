package elevatorStateMachine

import (
	"fmt"
	"time"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	hw "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
)

const (
	idle int = iota
	moving
	doorOpen
)

type Channels struct {
	OrderComplete chan int
	ElevatorState chan int
	StateError    chan error
	//Localqueue     chan [NumFloors][NumButtons]int
	NewOrderChan   chan Keypress
	ArrivedAtFloor chan int
	doorTimeout    chan bool
}

type elev struct {
	state int
	dir   Direction
	floor int
	queue [NumFloors][NumButtons]int
}

func ESM_loop(ch Channels, btnsPressed chan Keypress) {
	elevator := elev{
		state: idle,
		dir:   DirStop,
		floor: hw.GetFloorSensorSignal(),
		queue: [NumFloors][NumButtons]int{},
	}
	var doorTimedOut <-chan time.Time
	for {
		select {
		case newOrder := <-btnsPressed:
			// Everytime new order, update local orders
			// If idle, execute order
			fmt.Println("New order at floor ", newOrder.Floor, " for button ", newOrder.Btn)
			elevator.queue[newOrder.Floor][newOrder.Btn] = 1
			switch elevator.state {
			case idle:
				elevator.dir = chooseDirection(elevator)
				hw.SetMotorDirection(elevator.dir)
				if elevator.dir == DirStop {
					elevator.state = doorOpen
					doorTimedOut = time.After(3 * time.Second)
					hw.SetDoorOpenLamp(1)
					// TODO: This functionality should not be here. Should be in governor.
					// NB: Here we assume all orders are cleared at a floor.
					hw.SetButtonLamp(BtnUp, elevator.floor, 0)
					hw.SetButtonLamp(BtnDown, elevator.floor, 0)
					hw.SetButtonLamp(BtnInside, elevator.floor, 0)
					elevator.queue[elevator.floor][BtnUp] = 0
					elevator.queue[elevator.floor][BtnDown] = 0
					elevator.queue[elevator.floor][BtnInside] = 0
				} else {
					elevator.state = moving
				}
			case moving:
			case doorOpen:
				if chooseDirection(elevator) == DirStop {
					doorTimedOut = time.After(3 * time.Second)
				}
				// TODO: This functionality should not be here. Should be in governor.
				// NB: Here we assume all orders are cleared at a floor.
				hw.SetButtonLamp(BtnUp, elevator.floor, 0)
				hw.SetButtonLamp(BtnDown, elevator.floor, 0)
				hw.SetButtonLamp(BtnInside, elevator.floor, 0)
				elevator.queue[elevator.floor][BtnUp] = 0
				elevator.queue[elevator.floor][BtnDown] = 0
				elevator.queue[elevator.floor][BtnInside] = 0
			default:
				fmt.Println("default error")
			}
		case elevator.floor = <-ch.ArrivedAtFloor:
			fmt.Println("Arrived at floor", elevator.floor+1)
			//fmt.Println(elevator.queue)
			// Everytime we stop at floor, check if order complete
			// if order complete, stop elevator etc.
			if shouldStop(elevator) {
				hw.SetMotorDirection(DirStop)
				doorTimedOut = time.After(3 * time.Second)
				elevator.state = doorOpen
				// TODO: This functionality should not be here. Should be in governor.
				// NB: Here we assume all orders are cleared at a floor.
				hw.SetButtonLamp(BtnUp, elevator.floor, 0)
				hw.SetButtonLamp(BtnDown, elevator.floor, 0)
				hw.SetButtonLamp(BtnInside, elevator.floor, 0)
				hw.SetDoorOpenLamp(1)
				elevator.queue[elevator.floor][BtnUp] = 0
				elevator.queue[elevator.floor][BtnDown] = 0
				elevator.queue[elevator.floor][BtnInside] = 0
			}
		case <-doorTimedOut:
			// Order complete, send confirmation
			// check if any more orders, go to idle if not
			// TODO: send order complete to governor
			fmt.Println("Door timeout")
			hw.SetDoorOpenLamp(0)
			elevator.dir = chooseDirection(elevator)
			if elevator.dir == DirStop {
				elevator.state = idle
			} else {
				elevator.state = moving
				hw.SetMotorDirection(elevator.dir)
			}
		}
	}
}

// NOTE: Might be usual functionality, left here as an example
// go func() {
// ch.doorTimeout <- true
// }()

func shouldStop(elevator elev) bool {
	switch elevator.dir {
	case DirUp:
		return elevator.queue[elevator.floor][BtnUp] == 1 ||
			elevator.queue[elevator.floor][BtnInside] == 1 ||
			!ordersAbove(elevator)
	case DirDown:
		return elevator.queue[elevator.floor][BtnDown] == 1 ||
			elevator.queue[elevator.floor][BtnInside] == 1 ||
			!ordersBelow(elevator)
	case DirStop:
	default:
		fmt.Println("something went wrong with shouldStop")
	}
	return false
}

func chooseDirection(elevator elev) Direction {
	switch elevator.dir {
	case DirStop:
		if ordersAbove(elevator) {
			return DirUp
		} else if ordersBelow(elevator) {
			return DirDown
		} else {
			return DirStop
		}
	case DirUp:
		if ordersAbove(elevator) {
			return DirUp
		} else if ordersBelow(elevator) {
			return DirDown
		} else {
			return DirStop
		}

	case DirDown:
		if ordersBelow(elevator) {
			return DirDown
		} else if ordersAbove(elevator) {
			return DirUp
		} else {
			return DirStop
		}
	}
	return DirStop
}

func ordersAbove(elevator elev) bool {
	for floor := elevator.floor + 1; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if elevator.queue[floor][btn] == 1 {
				return true
			}
		}
	}
	return false
}

func ordersBelow(elevator elev) bool {
	for floor := 0; floor < elevator.floor; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if elevator.queue[floor][btn] == 1 {
				return true
			}
		}
	}
	return false
}
