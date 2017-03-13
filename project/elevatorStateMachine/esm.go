package elevatorStateMachine

import (
	"fmt"
	"time"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	hw "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
)

const (
	undefined int = iota - 1
	idle
	moving
	doorOpen
)

type Channels struct {
	OrderComplete  chan int
	ElevatorChan   chan Elev
	StateError     chan error
	NewOrderChan   chan Keypress
	ArrivedAtFloor chan int
}

func ESM_loop(ch Channels, btnsPressed chan Keypress) {
	elevator := Elev{
		State: idle,
		Dir:   DirStop,
		Floor: hw.GetFloorSensorSignal(),
		Queue: [NumFloors][NumButtons]bool{},
	}
	engineErrorTimer := time.NewTimer(3 * time.Second)
	engineErrorTimer.Stop()
	var doorTimedOut <-chan time.Time
	ch.ElevatorChan <- elevator
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
					hw.SetDoorOpenLamp(1)
					doorTimedOut = time.After(3 * time.Second)
					// NB: Here we assume all orders are cleared at a floor.
					go func() { ch.OrderComplete <- newOrder.Floor }()
					elevator.Queue[elevator.Floor] = [NumButtons]bool{}
				} else {
					elevator.State = moving
					engineErrorTimer.Reset(3 * time.Second)
				}

			case moving:
			case doorOpen:
				if elevator.Floor == newOrder.Floor {
					doorTimedOut = time.After(3 * time.Second)
					go func() { ch.OrderComplete <- newOrder.Floor }()
					// NB: Here we assume all orders are cleared at a floor.
					elevator.Queue[elevator.Floor] = [NumButtons]bool{}
				}

			case undefined:
			default:
				fmt.Println("idledefault error")
			}
			ch.ElevatorChan <- elevator

		case elevator.Floor = <-ch.ArrivedAtFloor:
			if elevator.State == undefined {
				//trigg peer update enable osv.
			}
			fmt.Println("Arrived at floor", elevator.Floor+1)
			if shouldStop(elevator) {
				hw.SetMotorDirection(DirStop)
				doorTimedOut = time.After(3 * time.Second)
				elevator.State = doorOpen
				hw.SetDoorOpenLamp(1)
				// NB: This clears all orders on the given floor
				elevator.Queue[elevator.Floor] = [NumButtons]bool{}
				go func() { ch.OrderComplete <- elevator.Floor }()
				engineErrorTimer.Stop()
			} else {
				engineErrorTimer.Reset(3 * time.Second)
			}
			ch.ElevatorChan <- elevator

		case <-doorTimedOut:
			hw.SetDoorOpenLamp(0)
			elevator.Dir = chooseDirection(elevator)
			if elevator.Dir == DirStop {
				elevator.State = idle
			} else {
				elevator.State = moving
				engineErrorTimer.Reset(3 * time.Second)
				hw.SetMotorDirection(elevator.Dir)
			}
		case <-engineErrorTimer.C:
			// QUESTION: Do we need to handle special case of eg. not at same floor || sensorSignal==-1 ?
			hw.SetMotorDirection(DirStop)
			elevator.State = undefined
			fmt.Println("\x1b[31;1m", "MOTOR STOP: Initiate precausionary measures!", "\x1b[0m")
			//peers.transmitter disable
			for i := 0; i < 10; i++ {
				if i%2 == 0 {
					hw.SetStopLamp(1)
				} else {
					hw.SetStopLamp(0)
				}
				time.Sleep(time.Millisecond * 200)
			}
			hw.SetMotorDirection(elevator.Dir)
			ch.ElevatorChan <- elevator
			engineErrorTimer.Reset(5 * time.Second)
			fmt.Println("nå")
		}
	}
}

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
