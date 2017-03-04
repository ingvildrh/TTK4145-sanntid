package elevatorGovernor

import (
	"fmt"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	esm "github.com/perkjelsvik/TTK4145-sanntid/project/elevatorStateMachine"
	hw "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
)

var orderTimeout chan bool

var elevList [NumElevators]Elev

//type btnList [NumButtons]bool

//var syncBtnLights chan bool
//var queue [NumFloors][NumElevators]btnList
//var elevStates [NumElevators]elevStatus
var id int

//NOTE: queue and state info suggestion so far
/*  #1		 #2		  #3
floor  floor  floor
 dir		 dir		dir
state	state	 state
0 0 0  0 0 0  0 0 0
0 0 0  0 0 0  0 0 0
0 0 0  0 0 0  0 0 0
0 0 0  0 0 0  0 0 0  ] 0 = false, 1 = true
*/

/*
func GOV_init(ID int, ch esm.Channels, btnsPressed chan Keypress) {
	hw.SetStopLamp(1)
	var queue = [NumFloors][NumButtons][NumElevators]int{}
	fmt.Println(queue)
	var press Keypress

	for floor := 0; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if queue[floor][ID][btn] == 1 {}
				press.Floor = floor
				press.Btn = Button(btn)
				ch.NewOrderChan <- press
			}
		}
	}

	go GOV_loop(ID, ch)
	go GOV_lightsLoop()

}*/

// TODO: Deal with elevatorState and StateError channels
func GOV_loop(ID int, ch esm.Channels, btnsPressed chan Keypress, NetworkUpdate chan int, syncBtnLights chan bool) {
	id = ID
	var designatedElevator int
	for {
		select {
		case newLocalOrder := <-btnsPressed:
			fmt.Println("New order at floor ", newLocalOrder.Floor, " for button ", newLocalOrder.Btn)
			if !duplicateOrder(newLocalOrder) {
				designatedElevator = costCalculator(newLocalOrder)
				fmt.Println("new local order given to: ", designatedElevator)
				elevList[id].Queue[newLocalOrder.Floor][newLocalOrder.Btn] = true
				ch.NewOrderChan <- newLocalOrder
				syncBtnLights <- true
				//networkUpdate <- {newLocalOrder, designatedElevator}
			}
			// QUESTION: can we push to NetworkUpdate from first case without triggering this case? Or do we need two channels!?
		case completedFloor := <-ch.OrderComplete:
			elevList[id].Queue[completedFloor] = [NumButtons]bool{}
			syncBtnLights <- true
		case elevList[id] = <-ch.ElevatorChan:

		/*
			tempStates[id] = <- ElevatorState:
			elevStates.state = tempStates[id][2]
		*/
		case <-NetworkUpdate:
			syncBtnLights <- true
		}
	}
}

func duplicateOrder(order Keypress) bool {
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

func costCalculator(order Keypress) int {
	//FIXME: This cost calcultor is stupid
	elevList[1].Floor = 3
	elevList[2].Floor = 2
	fmt.Println("etasje nummer ", elevList[id].Floor)
	floorDiff := 0
	minCost := 10
	bestElevator := id
	for elevator := 0; elevator < NumElevators; elevator++ {
		// QUESTION: How to do Abs() properly?? any way?
		floorDiff = order.Floor - elevList[elevator].Floor
		if floorDiff == 0 {
			return elevator
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

func GOV_lightsLoop(syncBtnLights chan bool) {
	for {
		<-syncBtnLights
		for floor := 0; floor < NumFloors; floor++ {
			for btn := BtnUp; btn < NumButtons; btn++ {
				if elevList[id].Queue[floor][btn] {
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
