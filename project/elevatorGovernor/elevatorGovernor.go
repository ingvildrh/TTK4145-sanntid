package elevatorGovernor

import (
	"fmt"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/config"
	hw "github.com/perkjelsvik/TTK4145-sanntid/project/hardware"
)

// Structure of elevList (0 = Inactive, 1 = Active)
/*------------------------*
|    id1		id2		 id3		|
| 	state  state  state		|
|    dir		dir		 dir		|
| 	floor	 floor	floor		|
| 	0 0 0  0 0 0  0 0 0		|
| 	0 0 0  0 0 0  0 0 0		|
| 	0 0 0  0 0 0  0 0 0		|
| 	0 0 0  0 0 0  0 0 0		|
*------------------------*/

// Governate called as goroutine; receives button presses, assigns esm and updates sync
func Governate(id int, btnsPressedCh chan Keypress, lightUpdateCh chan [NumElevators]Elev,
	orderCompleteCh chan int, newOrderCh chan Keypress, elevatorCh chan Elev,
	orderUpdateCh chan Keypress, updateSyncCh chan Elev, updateGovernorCh chan [NumElevators]Elev,
	onlineElevatorsCh chan [NumElevators]bool) {

	var (
		elevList       [NumElevators]Elev
		onlineList     [NumElevators]bool
		completedOrder Keypress
	)
	completedOrder.DesignatedElevator = id
	elevList[id] = <-elevatorCh
	updateSyncCh <- elevList[id]

	for {
		select {
		case newLocalOrder := <-btnsPressedCh:
			if !onlineList[id] && newLocalOrder.Btn == BtnInside {
				elevList[id].Queue[newLocalOrder.Floor][BtnInside] = true
				lightUpdateCh <- elevList
				go func() { newOrderCh <- newLocalOrder }()
			} else if !onlineList[id] && newLocalOrder.Btn != BtnInside {
				// Do nothing
				continue
			} else {
				if newLocalOrder.Floor == elevList[id].Floor && elevList[id].State != Moving {
					newOrderCh <- newLocalOrder
				} else {
					if !duplicateOrder(newLocalOrder, elevList, id) {
						fmt.Println("New order at floor ", newLocalOrder.Floor+1, " for button ", newLocalOrder.Btn)
						newLocalOrder.DesignatedElevator = costCalculator(newLocalOrder, elevList, id, onlineList)
						orderUpdateCh <- newLocalOrder
					}
				}
			}

		case completedOrder.Floor = <-orderCompleteCh:
			completedOrder.Done = true
			for btn := BtnUp; btn < NumButtons; btn++ {
				if elevList[id].Queue[completedOrder.Floor][btn] {
					completedOrder.Btn = btn
				}
				for elevator := 0; elevator < NumElevators; elevator++ {
					if btn != BtnInside || elevator == id {
						elevList[elevator].Queue[completedOrder.Floor][btn] = false
					}
				}
			}

			if onlineList[id] {
				orderUpdateCh <- completedOrder
			}
			lightUpdateCh <- elevList

		case newElev := <-elevatorCh:
			tmpQueue := elevList[id].Queue
			if elevList[id].State == Undefined && newElev.State != Undefined {
				onlineList[id] = true
			}
			elevList[id] = newElev
			elevList[id].Queue = tmpQueue
			if onlineList[id] {
				updateSyncCh <- elevList[id]
			}

		case copyOnlineList := <-onlineElevatorsCh:
			onlineList = copyOnlineList

		case tmpElevList := <-updateGovernorCh:
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
					if tmpElevList[id].Queue[floor][btn] && !elevList[id].Queue[floor][btn] {
						elevList[id].Queue[floor][btn] = true
						order := Keypress{Floor: floor, Btn: btn, DesignatedElevator: id, Done: false}
						go func() { newOrderCh <- order }()
						newOrder = true
					} else if !tmpElevList[id].Queue[floor][btn] && elevList[id].Queue[floor][btn] {
						elevList[id].Queue[floor][btn] = false
						order := Keypress{Floor: floor, Btn: btn, DesignatedElevator: id, Done: true}
						go func() { newOrderCh <- order }()
						newOrder = true
					}
				}
			}

			if newOrder {
				lightUpdateCh <- elevList
			}
		}
	}
}

// LightSetter called as goroutine, responsible for setting/clearing lights for every order change
func LightSetter(lightUpdateChan <-chan [NumElevators]Elev, id int) {
	var orderExists [NumElevators]bool

	for {
		elevs := <-lightUpdateChan
		for floor := 0; floor < NumFloors; floor++ {
			for btn := BtnUp; btn < NumButtons; btn++ {
				for elevator := 0; elevator < NumElevators; elevator++ {
					orderExists[elevator] = false
					if elevator != id && btn == BtnInside {
						// Ignore inside orders for other elevators
						continue
					}
					if elevs[elevator].Queue[floor][btn] {
						hw.SetButtonLamp(btn, floor, 1)
						orderExists[elevator] = true
					}
				}
				if orderExists == [NumElevators]bool{} {
					hw.SetButtonLamp(btn, floor, 0)
				}
			}
		}
	}
}
