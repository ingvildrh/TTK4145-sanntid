package elevatorGovernor

import (
	"fmt"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	esm "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorStateMachine"
	hw "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
)

//NOTE: queue and state info suggestion so far
/*
  #1		 #2		  #3
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
	designatedElevator := id
	var completedOrder Keypress
	completedOrder.DesignatedElevator = id
	for {
		select {
		//QUESTION: burde vi flytte btnsPressed til Sync?? hehe
		case newLocalOrder := <-btnsPressed:
			if !duplicateOrder(newLocalOrder, elevList, id) {
				fmt.Println("New order at floor ", newLocalOrder.Floor, " for button ", newLocalOrder.Btn)
				newLocalOrder.DesignatedElevator = costCalculator(newLocalOrder, elevList, id)
				fmt.Println("new local order given to: ", designatedElevator)
				orderUpdate <- newLocalOrder
			}
		case completedOrder.Floor = <-ch.OrderComplete:
			completedOrder.Done = true
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
						fmt.Println(elevList[id].Queue[floor])
						order := Keypress{Floor: floor, Btn: btn, DesignatedElevator: id, Done: false}
						ch.NewOrderChan <- order
						orderB := <-ch.OrderComplete
						fmt.Println("BTN SENT: ", btn)
						fmt.Println(orderB)
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
	elevList[1].Floor = 3
	elevList[2].Floor = 2
	fmt.Println("etasje nummer ", elevList[id].Floor)
	minCost := 10
	bestElevator := id
	for elevator := 0; elevator < NumElevators; elevator++ {
		// QUESTION: How to do Abs() properly?? any way?
		floorDiff := order.Floor - elevList[elevator].Floor
		if floorDiff == 0 {
			return id
		} else if floorDiff < 0 {
			floorDiff = -floorDiff
		}

		if minCost > floorDiff {
			minCost = floorDiff
			bestElevator = elevator
		}

	}
	fmt.Println("BEST: ", bestElevator)
	return id
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

/*NOTE: should we have a compare between new network queue
and already existing queue? only forward new orders to esm
if actually new order? Could be done by for example:
newQueue[floor][btn] != queue[floor][btn]
*/
