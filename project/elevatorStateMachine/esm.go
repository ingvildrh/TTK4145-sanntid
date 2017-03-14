package elevatorStateMachine

import (
	"fmt"
	"time"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/config"
	hw "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
)

// StateMachineChannels contains all channels between governor - esm and hardware - esm
type StateMachineChannels struct {
	OrderComplete  chan int
	Elevator       chan Elev
	StateError     chan error
	NewOrder       chan Keypress
	ArrivedAtFloor chan int
}

// RunElevator called as a goroutine; runs elevator and updates governor for changes
func RunElevator(ch StateMachineChannels) {
	elevator := Elev{
		State: Idle,
		Dir:   DirStop,
		Floor: hw.GetFloorSensorSignal(),
		Queue: [NumFloors][NumButtons]bool{},
	}
	engineErrorTimer := time.NewTimer(3 * time.Second)
	engineErrorTimer.Stop()
	orderCleared := false
	doorTimedOut := time.NewTimer(3 * time.Second)
	ch.Elevator <- elevator

	for {
		select {
		case newOrder := <-ch.NewOrder:
			if newOrder.Done {
				elevator.Queue[newOrder.Floor][BtnUp] = false
				elevator.Queue[newOrder.Floor][BtnDown] = false
				orderCleared = true
			} else {
				elevator.Queue[newOrder.Floor][newOrder.Btn] = true
			}
			switch elevator.State {

			case Idle:
				elevator.Dir = chooseDirection(elevator)
				hw.SetMotorDirection(elevator.Dir)
				if elevator.Dir == DirStop {
					elevator.State = DoorOpen
					hw.SetDoorOpenLamp(1)
					doorTimedOut.Reset(3 * time.Second)
					go func() { ch.OrderComplete <- newOrder.Floor }()
					elevator.Queue[elevator.Floor] = [NumButtons]bool{}
				} else {
					elevator.State = Moving
					engineErrorTimer.Reset(3 * time.Second)
				}

			case Moving:
			case DoorOpen:
				if elevator.Floor == newOrder.Floor {
					doorTimedOut.Reset(3 * time.Second)
					go func() { ch.OrderComplete <- newOrder.Floor }()
					elevator.Queue[elevator.Floor] = [NumButtons]bool{}
				}

			case Undefined:
			default:
				fmt.Println("Idle default error")
			}
			ch.Elevator <- elevator

		case elevator.Floor = <-ch.ArrivedAtFloor:
			fmt.Println("Arrived at floor", elevator.Floor+1)
			if shouldStop(elevator) ||
				(!shouldStop(elevator) && elevator.Queue == [NumFloors][NumButtons]bool{} && orderCleared) {
				// NB: Test without orderCleared, shouldn't be necessary
				orderCleared = false
				hw.SetMotorDirection(DirStop)
				hw.SetDoorOpenLamp(1)
				engineErrorTimer.Stop()
				elevator.State = DoorOpen
				doorTimedOut.Reset(3 * time.Second)
				elevator.Queue[elevator.Floor] = [NumButtons]bool{}
				go func() { ch.OrderComplete <- elevator.Floor }()
			} else if elevator.State == Moving {
				engineErrorTimer.Reset(3 * time.Second)
			}
			ch.Elevator <- elevator

		case <-doorTimedOut.C:
			hw.SetDoorOpenLamp(0)
			elevator.Dir = chooseDirection(elevator)
			if elevator.Dir == DirStop {
				elevator.State = Idle
				engineErrorTimer.Stop()
			} else {
				elevator.State = Moving
				engineErrorTimer.Reset(3 * time.Second)
				hw.SetMotorDirection(elevator.Dir)
			}
			ch.Elevator <- elevator

		case <-engineErrorTimer.C:
			hw.SetMotorDirection(DirStop)
			elevator.State = Undefined
			fmt.Println("\x1b[1;1;33m", "MOTOR STOP: Initiate precausionary measures!", "\x1b[0m")
			for i := 0; i < 10; i++ {
				if i%2 == 0 {
					hw.SetStopLamp(1)
				} else {
					hw.SetStopLamp(0)
				}
				time.Sleep(time.Millisecond * 200)
			}
			hw.SetMotorDirection(elevator.Dir)
			ch.Elevator <- elevator
			engineErrorTimer.Reset(5 * time.Second)
		}
	}
}
