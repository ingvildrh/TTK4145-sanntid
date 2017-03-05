package syncElevators

import (
	"fmt"
	"time"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
)

type SyncChannels struct {
	UpdateGovernor   chan [NumElevators]Elev
	UpdateSync       chan Elev
	OrderUpdate      chan Keypress
	IncomingMsg      chan Message
	OutgoingMsg      chan Message
	updatePeersTimer chan time.Time
}

//QUESTION: should we ACK the ACK? Timeout the ACK? Or simply CheckAgain if one or more elvators become offline
/*
									ACK MATRIX
{BtnUp BtnDown DesignatedElevator elev1 elev2 elev3}
{BtnUp BtnDown DesignatedElevator elev1 elev2 elev3}
{BtnUp BtnDown DesignatedElevator elev1 elev2 elev3}
{BtnUp BtnDown DesignatedElevator elev1 elev2 elev3}
*/

func SYNC_loop(ch SyncChannels, id int) {

	var registeredOrders [NumFloors]AckMatrix
	var elevList [NumElevators]Elev
	var sendMsg Message
	var allAcked [NumElevators]Acknowledge
	//NOTE: allAcked := [NumElevators]Acknowledge{Acked, Acked, Acked}
	for i := 0; i < NumElevators; i++ {
		allAcked[i] = Acked
	}
	var updatePeersTimer <-chan time.Time
	//updatePeersTimer = time.After(100 * time.Millisecond)
	var designatedElevator int
	// NOTE: burde vi importere constants som def eller liknende? mer lesbart
	for {
		select {
		case tmpElev := <-ch.UpdateSync:
			tmpQueue := elevList[id].Queue
			elevList[id] = tmpElev
			elevList[id].Queue = tmpQueue
		case newOrder := <-ch.OrderUpdate:
			if newOrder.Done {
				registeredOrders[newOrder.Floor].ImplicitAcks[id] = Finished
				elevList[id].Queue[newOrder.Floor] = [NumButtons]bool{}
			} else {
				if newOrder.Btn == BtnInside {
					elevList[id].Queue[newOrder.Floor][newOrder.Btn] = true
				} else {
					registeredOrders[newOrder.Floor].DesignatedElevator = newOrder.DesignatedElevator
					//NB: this is for testing purposes
					registeredOrders[newOrder.Floor].ImplicitAcks = allAcked
					if newOrder.Btn == BtnUp {
						registeredOrders[newOrder.Floor].OrderUp = true
					} else if newOrder.Btn == BtnDown {
						registeredOrders[newOrder.Floor].OrderDown = true
					}
					fmt.Println("BTN: ", registeredOrders[newOrder.Floor])
				}
			}
		case msg := <-ch.IncomingMsg:
			// IDEA: Have another ack-state ackButNotAllAcked.
			for elevator := 0; elevator < NumElevators; elevator++ {
				if elevator == id {
					continue
				}
				for floor := 0; floor < NumFloors; floor++ {
					if msg.RegisteredOrders[floor].ImplicitAcks[elevator] == Finished {
						registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id)
						// QUESTION: This might not be safe - what about internal orders and costCalculator?
						elevList[elevator].Queue[floor] = [NumButtons]bool{}
					}
					if msg.RegisteredOrders[floor].ImplicitAcks[elevator] == Acked {
						if registeredOrders[floor].ImplicitAcks[id] == Finished {
							registeredOrders[floor].ImplicitAcks[elevator] = msg.RegisteredOrders[floor].ImplicitAcks[elevator]
						} else {
							registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id)
						}
					}
					if registeredOrders[floor].ImplicitAcks == allAcked {
						designatedElevator = registeredOrders[floor].DesignatedElevator
						if registeredOrders[floor].OrderUp && registeredOrders[floor].OrderDown {
							elevList[designatedElevator].Queue[floor][BtnUp] = true
							elevList[designatedElevator].Queue[floor][BtnDown] = true
						} else if registeredOrders[floor].OrderUp {
							elevList[designatedElevator].Queue[floor][BtnUp] = true
						} else {
							elevList[designatedElevator].Queue[floor][BtnDown] = true
						}
						ch.UpdateGovernor <- elevList
						registeredOrders[floor].OrderUp = false
						registeredOrders[floor].OrderDown = false
					}
				}
			}
		case <-updatePeersTimer:
			sendMsg.RegisteredOrders = registeredOrders
			sendMsg.Elevator = elevList
			ch.OutgoingMsg <- sendMsg
			updatePeersTimer = time.After(100 * time.Millisecond)
		}
	}
}

func copyMessage(msg Message, registeredOrders [NumFloors]AckMatrix, elevator int, floor int, id int) [NumFloors]AckMatrix {
	registeredOrders[floor].ImplicitAcks[id] = msg.RegisteredOrders[floor].ImplicitAcks[elevator]
	registeredOrders[floor].ImplicitAcks[elevator] = msg.RegisteredOrders[floor].ImplicitAcks[elevator]
	registeredOrders[floor].DesignatedElevator = msg.RegisteredOrders[floor].DesignatedElevator
	registeredOrders[floor].OrderDown = msg.RegisteredOrders[floor].OrderDown
	registeredOrders[floor].OrderUp = msg.RegisteredOrders[floor].OrderUp
	return registeredOrders
}
