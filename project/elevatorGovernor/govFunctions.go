package elevatorGovernor

import (
	"fmt"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/config"
)

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
	if order.Btn == BtnInside {
		return id
	}
	minCost := (NumButtons * NumFloors) * NumElevators
	bestElevator := id
	for elevator := 0; elevator < NumElevators; elevator++ {
		if !onlineList[elevator] {
			// Disregarding offline elevators
			continue
		}
		cost := order.Floor - elevList[elevator].Floor

		if cost == 0 && elevList[elevator].State != Moving {
			fmt.Println("Assigned elevator: ", bestElevator)
			fmt.Println("Order cost was: ", cost)
			bestElevator = elevator
			return bestElevator
		}
		if cost < 0 {
			cost = -cost
			if elevList[elevator].Dir == DirUp {
				fmt.Println("DIR UP")
				cost += 3
			}
		} else if cost > 0 {
			if elevList[elevator].Dir == DirDown {
				fmt.Println("DIR DOWN")
				cost += 3
			}
		}

		if elevList[elevator].State == DoorOpen {
			cost++
		}

		if cost < minCost {
			minCost = cost
			bestElevator = elevator
		}

		fmt.Println("elevator ", elevator, "has cost ", cost)
		fmt.Println("and is in floor ", elevList[elevator].Floor+1)
	}
	fmt.Println("Assigned elevator: ", bestElevator)
	fmt.Println("Order cost was", minCost)
	return bestElevator
}
