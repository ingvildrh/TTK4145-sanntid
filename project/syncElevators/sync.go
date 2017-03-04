package syncElevators

import (
	"time"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
)

var registeredOrders [NumFloors]AckMatrix
var elevList [NumElevators]Elev

//QUESTION: should we ACK the ACK? Timeout the ACK? Or simply CheckAgain if one or more elvators become offline
/*
									ACK MATRIX
{BtnUp BtnDown DesignatedElevator elev1 elev2 elev3}
{BtnUp BtnDown DesignatedElevator elev1 elev2 elev3}
{BtnUp BtnDown DesignatedElevator elev1 elev2 elev3}
{BtnUp BtnDown DesignatedElevator elev1 elev2 elev3}
*/

func SYNC_loop(NetworkUpdate chan int, incomingMsg chan Msg, id int) {
	allAcked := [NumElevators]Acknowledge{Acked, Acked, Acked}
	var designatedElevator int
	var btn Button
	// NOTE: burde vi importere constants som def eller liknende? mer lesbart
	for {
		select {
		case <-NetworkUpdate:

		case msg := <-incomingMsg:
			for elevator := 0; elevator < NumElevators; elevator++ {
				if elevator == id {
					elevator++
				}
				for floor := 0; floor < NumFloors; floor++ {
					if msg.RegisteredOrders[floor].ImplicitAcks[elevator] == Finished {
						registeredOrders[floor].ImplicitAcks[id] = Finished
					}
					if msg.RegisteredOrders[floor].ImplicitAcks[elevator] == Acked {
						registeredOrders[floor].ImplicitAcks[id] = Acked
						registeredOrders[floor].ImplicitAcks[elevator] = Acked
					}
					if registeredOrders[floor].ImplicitAcks == allAcked {
						designatedElevator = registeredOrders[floor].DesignatedElevator
						if registeredOrders[floor].OrderUp && registeredOrders[floor].OrderDown {
							elevList[designatedElevator].Queue[floor][BtnUp] = true
							elevList[designatedElevator].Queue[floor][BtnDown] = true
						} else if registeredOrders[floor].OrderUp {
							btn = BtnUp
						} else {
							btn = BtnDown
						}
						elevList[designatedElevator].Queue[floor][btn] = true
					}
				}
			}
		}
	}
}

func SYNC_periodicStateSpammer(outgoingMsg chan Msg) {
	var message Msg
	for {
		time.Sleep(time.Millisecond * 100)
		message.RegisteredOrders = registeredOrders
		message.Elevator = elevList
		outgoingMsg <- message
	}
}
