package elevatorGovernor

import (
	"fmt"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	esm "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorStateMachine"
	hw "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
)

//NOTE: queue and state info suggestion so far
/*
  id1		 id2		 id3
 floor  floor  floor
  dir		dir		 dir
 state	 state	state
 0 0 0  0 0 0  0 0 0
 0 0 0  0 0 0  0 0 0
 0 0 0  0 0 0  0 0 0
 0 0 0  0 0 0  0 0 0
 0 = false, 1 = true
*/

// TODO: Deal with elevatorState and StateError channels
func GOV_loop(ID int, ch esm.Channels, btnsPressed chan Keypress,
	updateSync chan Elev, updateGovernor chan [NumElevators]Elev,
	orderUpdate chan Keypress, syncBtnLights chan [NumFloors][NumButtons]bool) {
	//var orderTimeout chan bool
	var elevList [NumElevators]Elev
	id := ID
	moving := 1
	//designatedElevator := id
	var completedOrder Keypress
	completedOrder.DesignatedElevator = id
	for {
		select {
		//QUESTION: burde vi flytte btnsPressed til Sync?? hehe
		case newLocalOrder := <-btnsPressed:
			// QUESTION: Move state: idle, moving and doorOpen to constants? Or something like this?
			if newLocalOrder.Floor == elevList[id].Floor && elevList[id].State != moving {
				ch.NewOrderChan <- newLocalOrder
			} else {
				if !duplicateOrder(newLocalOrder, elevList, id) {
					fmt.Println("New order at floor ", newLocalOrder.Floor+1, " for button ", PrintBtn(newLocalOrder.Btn))
					newLocalOrder.DesignatedElevator = costCalculator(newLocalOrder, elevList, id)
					//fmt.Println("new local order given to: ", designatedElevator)
					orderUpdate <- newLocalOrder
				}
			}

		case completedOrder.Floor = <-ch.OrderComplete:
			completedOrder.Done = true
			// QUESTION: We only return the floor. Here we set only 1 btnPress. Still acking works in sync?????????
			for btn := BtnUp; btn < NumButtons; btn++ {
				if elevList[id].Queue[completedOrder.Floor][btn] {
					completedOrder.Btn = btn
				}
			}
			elevList[id].Queue[completedOrder.Floor] = [NumButtons]bool{}
			syncBtnLights <- elevList[id].Queue
			orderUpdate <- completedOrder

		case tmpElev := <-ch.ElevatorChan:
			tmpQueue := elevList[id].Queue
			elevList[id] = tmpElev
			elevList[id].Queue = tmpQueue
			updateSync <- elevList[id]

		case tmpElevList := <-updateGovernor:
			newOrder := false
			for elevator := 0; elevator < NumElevators; elevator++ {
				if elevator == id {
					continue
				}
				elevList[elevator] = tmpElevList[elevator]
			}
			for btn := BtnUp; btn < NumButtons; btn++ {
				for floor := 0; floor < NumFloors; floor++ {
					// NOTE: potential problem of overwriting finished orders, then preventing new orders while acking finished
					if tmpElevList[id].Queue[floor][btn] && !elevList[id].Queue[floor][btn] {
						elevList[id].Queue[floor][btn] = true
						// NOTE: We don't really need to define DesignatedElevator since esm doesn't care
						order := Keypress{Floor: floor, Btn: btn, DesignatedElevator: id, Done: false}
						ch.NewOrderChan <- order
						newOrder = true
					}
				}
			}
			if newOrder {
				syncBtnLights <- elevList[id].Queue
			}
		}
	}
}

func duplicateOrder(order Keypress, elevList [NumElevators]Elev, id int) bool {
	if order.Btn == BtnInside && elevList[id].Queue[order.Floor][BtnInside] {
		return true
	}
	for elevator := 0; elevator < NumElevators; elevator++ {
		if elevList[id].Queue[order.Floor][order.Btn] {
			return true
		}
	}
	return false
}

func costCalculator(order Keypress, elevList [NumElevators]Elev, id int) int {
	//FIXME: This cost calcultor is stupid
	//elevList[1].Floor = 3
	//elevList[2].Floor = 2
	minCost := 10
	bestElevator := id
	for elevator := 0; elevator < NumElevators; elevator++ {
		// QUESTION: How to do Abs() properly?? any way?
		floorDiff := order.Floor - elevList[elevator].Floor
		if floorDiff == 0 {
			return bestElevator
		} else if floorDiff < 0 {
			floorDiff = -floorDiff
		}

		if minCost > floorDiff {
			minCost = floorDiff
			bestElevator = elevator
		}

	}
	return bestElevator
}

func GOV_lightsLoop(syncBtnLights chan [NumFloors][NumButtons]bool) {
	for {
		queue := <-syncBtnLights
		for floor := 0; floor < NumFloors; floor++ {
			for btn := BtnUp; btn < NumButtons; btn++ {
				if queue[floor][btn] {
					hw.SetButtonLamp(btn, floor, 1)
				} else {
					hw.SetButtonLamp(btn, floor, 0)
				}
			}
		}
	}
}
