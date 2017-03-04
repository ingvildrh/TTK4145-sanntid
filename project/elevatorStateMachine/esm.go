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
	OrderComplete  chan int
	ElevatorChan   chan Elev
	StateError     chan error
	NewOrderChan   chan Keypress
	ArrivedAtFloor chan int
	doorTimeout    chan bool
}

func ESM_loop(ch Channels, btnsPressed chan Keypress) {
	elevator := Elev{
		State: idle,
		Dir:   DirStop,
		Floor: hw.GetFloorSensorSignal(),
		Queue: [NumFloors][NumButtons]bool{},
	}
	var doorTimedOut <-chan time.Time
	for {
		select {
		case newOrder := <-ch.NewOrderChan:
			elevator.Queue[newOrder.Floor][newOrder.Btn] = true
			switch elevator.State {
			case idle:
				elevator.Dir = chooseDirection(elevator)
				hw.SetMotorDirection(elevator.Dir)
				if elevator.Dir == DirStop {
					elevator.State = doorOpen
					doorTimedOut = time.After(3 * time.Second)
					hw.SetDoorOpenLamp(1)
					// NB: Here we assume all orders are cleared at a floor.
					ch.OrderComplete <- newOrder.Floor
					elevator.Queue[elevator.Floor] = [NumButtons]bool{}
				} else {
					elevator.State = moving
				}
				ch.ElevatorChan <- elevator
				//ElevatorState <- Elev.floor, Elev.dir, Elev.state
			case moving:
			case doorOpen:
				if elevator.Floor == newOrder.Floor {
					doorTimedOut = time.After(3 * time.Second)
					ch.OrderComplete <- newOrder.Floor
					// NB: Here we assume all orders are cleared at a floor.
					elevator.Queue[elevator.Floor] = [NumButtons]bool{}
				}
			default:
				fmt.Println("default error")
			}
		case elevator.Floor = <-ch.ArrivedAtFloor:
			fmt.Println("Arrived at floor", elevator.Floor+1)
			if shouldStop(elevator) {
				hw.SetMotorDirection(DirStop)
				doorTimedOut = time.After(3 * time.Second)
				elevator.State = doorOpen
				hw.SetDoorOpenLamp(1)
				// NB: This clears all orders on the given floor
				elevator.Queue[elevator.Floor] = [NumButtons]bool{}
				ch.OrderComplete <- elevator.Floor
			}
			ch.ElevatorChan <- elevator
		case <-doorTimedOut:
			fmt.Println("Door timeout")
			hw.SetDoorOpenLamp(0)
			elevator.Dir = chooseDirection(elevator)
			if elevator.Dir == DirStop {
				elevator.State = idle
			} else {
				elevator.State = moving
				hw.SetMotorDirection(elevator.Dir)
			}
			ch.ElevatorChan <- elevator
		}
	}
}

// NOTE: Might be usual functionality, left here as an example
// go func() {
// ch.doorTimeout <- true
// }()

func shouldStop(elevator Elev) bool {
	switch elevator.Dir {
	case DirUp:
		return elevator.Queue[elevator.Floor][BtnUp] ||
			elevator.Queue[elevator.Floor][BtnInside] ||
			!ordersAbove(elevator)
	case DirDown:
		return elevator.Queue[elevator.Floor][BtnDown] ||
			elevator.Queue[elevator.Floor][BtnInside] ||
			!ordersBelow(elevator)
	case DirStop:
	default:
		fmt.Println("something went wrong with shouldStop")
	}
	return false
}

func chooseDirection(elevator Elev) Direction {
	switch elevator.Dir {
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

func ordersAbove(elevator Elev) bool {
	for floor := elevator.Floor + 1; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if elevator.Queue[floor][btn] {
				return true
			}
		}
	}
	return false
}

func ordersBelow(elevator Elev) bool {
	for floor := 0; floor < elevator.Floor; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if elevator.Queue[floor][btn] {
				return true
			}
		}
	}
	return false
}
