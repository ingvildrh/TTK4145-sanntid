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
 state  state  state
  dir		dir		 dir
 floor	 floor	floor
 0 0 0  0 0 0  0 0 0
 0 0 0  0 0 0  0 0 0
 0 0 0  0 0 0  0 0 0
 0 0 0  0 0 0  0 0 0
 0 = false, 1 = true
*/

// TODO: Deal with elevatorState and StateError channels
func GOV_loop(ID int, ch esm.Channels, orderUpdate chan Keypress, btnsPressed chan Keypress,
	updateSync chan Elev, updateGovernor chan [NumElevators]Elev,
	syncBtnLights chan [NumElevators]Elev, onlineElevators chan [NumElevators]bool) { //[NumFloors][NumButtons]bool) {
	//var orderTimeout chan bool
	var elevList [NumElevators]Elev
	var onlineList [NumElevators]bool
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
					newLocalOrder.DesignatedElevator = costCalculator(newLocalOrder, elevList, id, onlineList)
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
			//syncBtnLights <- elevList //[id].Queue
			orderUpdate <- completedOrder
			fmt.Println("We will clear light for", completedOrder.Floor+1, PrintBtn(completedOrder.Btn))
			fmt.Println()
			// NOTE: GOOD WAY TO PRINT THE QUEUES
			/*for f := NumFloors - 1; f > -1; f-- {
				fmt.Println("\t0: ", elevList[0].Queue[f], "\t1: ", elevList[1].Queue[f])
			}*/
			fmt.Println()
			syncBtnLights <- elevList //[id].Queue

		case tmpElev := <-ch.ElevatorChan:
			tmpQueue := elevList[id].Queue
			elevList[id] = tmpElev
			elevList[id].Queue = tmpQueue
			updateSync <- elevList[id]

		case tmpOnlineList := <-onlineElevators:
			onlineList = tmpOnlineList

		case tmpElevList := <-updateGovernor:
			//fmt.Println("Some change! Governator updated")
			newOrder := false
			for elevator := 0; elevator < NumElevators; elevator++ {
				if elevator == id {
					continue
				}
				if elevList[elevator].Queue != tmpElevList[elevator].Queue {
					newOrder = true
				}
				elevList[elevator] = tmpElevList[elevator]
			}
			for floor := 0; floor < NumFloors; floor++ {
				for btn := BtnUp; btn < NumButtons; btn++ {
					// NOTE: potential problem of overwriting finished orders, then preventing new orders while acking finished
					if tmpElevList[id].Queue[floor][btn] && !elevList[id].Queue[floor][btn] {
						elevList[id].Queue[floor][btn] = true
						// NOTE: We don't really need to define DesignatedElevator since esm doesn't care
						order := Keypress{Floor: floor, Btn: btn, DesignatedElevator: id, Done: false}
						go func() { ch.NewOrderChan <- order }()
						newOrder = true
					}
				}
			}
			if newOrder {
				syncBtnLights <- elevList
				//syncBtnLights <- elevList[elevator].Queue
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

func costCalculator(order Keypress, elevList [NumElevators]Elev, id int, onlineList [NumElevators]bool) int {
	//FIXME: This cost calcultor is stupid
	if order.Btn == BtnInside {
		return id
	}
	minCost := 32
	bestElevator := id
	//FIXME: should move to constnts, probably
	idle := 0
	moving := 1
	doorOpen := 2
	for elevator := 0; elevator < NumElevators; elevator++ {
		// QUESTION: How to do Abs() properly?? any way?
		//fmt.Println("Heis ", elevator, "er på etasje ", elevList[elevator].Floor+1)
		//fmt.Println("og den har state ", elevList[elevator].State)
		//fmt.Println("og den har Dir", elevList[elevator].Dir)
		fmt.Println("COST (online):", onlineList)
		if !onlineList[elevator] {
			continue //disregarding dead elevators
		}
		floorDiff := order.Floor - elevList[elevator].Floor
		cost := floorDiff
		if floorDiff == 0 && elevList[elevator].State != moving {
			fmt.Println("Assigned elevator: ", bestElevator)
			fmt.Println("Order cost was: ", cost)
			bestElevator = elevator
			return bestElevator
		}
		if elevList[elevator].Queue != [NumFloors][NumButtons]bool{} {
			cost += 2
		}
		if floorDiff < 0 {
			cost = -cost
			if elevList[elevator].Dir == DirUp {
				cost += 3
			}
		} else if floorDiff > 0 {
			if elevList[elevator].Dir == DirDown {
				cost += 3
			}
		}

		switch elevList[elevator].State {
		case doorOpen:
			cost += 3
		case idle:
			cost++
		case moving:
			cost++
		}
		if cost < minCost {
			minCost = cost
			bestElevator = elevator
		}
		fmt.Println("elevator ", elevator, "has cost ", cost)
	}
	fmt.Println("Assigned elevator: ", bestElevator)
	fmt.Println("Order cost was", minCost)
	return bestElevator
}

func GOV_lightsLoop(syncBtnLights chan [NumElevators]Elev, id int) {
	var (
		orderExists      [NumElevators]bool
		orderDoesntExist [NumElevators]bool
	)
	for {
		allQueues := <-syncBtnLights
		for floor := 0; floor < NumFloors; floor++ {
			for btn := BtnUp; btn < NumButtons; btn++ {
				for elevator := 0; elevator < NumElevators; elevator++ {
					orderExists[elevator] = false
					if elevator != id && btn == BtnInside {
						continue
					}
					if !allQueues[id].Queue[floor][btn] && btn == BtnInside {
						hw.SetButtonLamp(btn, floor, 0)
					}
					if allQueues[elevator].Queue[floor][btn] {
						hw.SetButtonLamp(btn, floor, 1)
						orderExists[elevator] = true
					}
				}
				if orderExists == orderDoesntExist {
					hw.SetButtonLamp(btn, floor, 0)
				}
			}
		}
	}
}
