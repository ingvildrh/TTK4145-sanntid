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
{assignedID elev1 elev2 elev3} {assignedID elev1 elev2 elev3}
{assignedID elev1 elev2 elev3} {assignedID elev1 elev2 elev3}
{assignedID elev1 elev2 elev3} {assignedID elev1 elev2 elev3}
{assignedID elev1 elev2 elev3} {assignedID elev1 elev2 elev3}
*/

func SYNC_loop(ch SyncChannels, id int) {

	fmt.Println("Sync loop started!")
	var registeredOrders [NumFloors][NumButtons - 1]AckList
	var elevList [NumElevators]Elev
	var sendMsg Message
	var allAcked [NumElevators]Acknowledge
	//NOTE: allAcked := [NumElevators]Acknowledge{Acked, Acked, Acked}
	for i := 0; i < NumElevators; i++ {
		allAcked[i] = Acked
	}
	var updatePeersTimer <-chan time.Time
	updatePeersTimer = time.After(100 * time.Millisecond)
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
				//registeredOrders[newOrder.Floor][newOrder.Btn].ImplicitAcks[id] = Finished
				elevList[id].Queue[newOrder.Floor] = [NumButtons]bool{}
				if newOrder.Btn != BtnInside {
					// FIXME: this is to prevent out of index because of BtnInside. Need better fix.
					registeredOrders[newOrder.Floor][newOrder.Btn].ImplicitAcks = [NumElevators]Acknowledge{NotAcked, NotAcked}
				}
			} else {
				if newOrder.Btn == BtnInside {
					// NB: Should probably send on net before adding to the queue. Exactly how unclear for now. To avoid immediate death after internal light on
					elevList[id].Queue[newOrder.Floor][newOrder.Btn] = true
				} else {
					registeredOrders[newOrder.Floor][newOrder.Btn].DesignatedElevator = 2
					registeredOrders[newOrder.Floor][newOrder.Btn].DesignatedElevator = newOrder.DesignatedElevator
					//NB: this is for testing purposes
					registeredOrders[newOrder.Floor][newOrder.Btn].ImplicitAcks = allAcked
				}
				// // sende intern knappebestilling tilbake!!
				// ch.UpdateGovernor <- elevList
			}
		case msg := <-ch.IncomingMsg:
			someChange := false
			fmt.Println("Hello from me")
			// IDEA: Have another ack-state ackButNotAllAcked.
			for elevator := 0; elevator < NumElevators; elevator++ {
				if elevator == id {
					continue
				}
				for floor := 0; floor < NumFloors; floor++ {
					for btn := BtnUp; btn < BtnInside; btn++ {
						if msg.RegisteredOrders[floor][btn].ImplicitAcks[elevator] == Finished {
							registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
							// QUESTION: This might not be safe - what about internal orders and costCalculator?
							elevList[elevator].Queue[floor] = [NumButtons]bool{}
							someChange = true
						}
						if msg.RegisteredOrders[floor][btn].ImplicitAcks[elevator] == Acked &&
							registeredOrders[floor][btn].ImplicitAcks[elevator] != Acked &&
							registeredOrders[floor][btn].ImplicitAcks[id] != Acked {
							someChange = true
							if registeredOrders[floor][btn].ImplicitAcks[id] == Finished {
								registeredOrders[floor][btn].ImplicitAcks[elevator] = msg.RegisteredOrders[floor][btn].ImplicitAcks[elevator]
							} else {
								registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
							}
						}
						if registeredOrders[floor][btn].ImplicitAcks == allAcked && !elevList[designatedElevator].Queue[floor][btn] {
							designatedElevator = registeredOrders[floor][btn].DesignatedElevator
							elevList[designatedElevator].Queue[floor][btn] = true
							someChange = true
						}
					}
				}
			}
			if someChange {
				ch.UpdateGovernor <- elevList
			}
		case <-updatePeersTimer:
			fmt.Println("Hello to you")
			sendMsg.RegisteredOrders = registeredOrders
			sendMsg.Elevator = elevList
			ch.OutgoingMsg <- sendMsg
			updatePeersTimer = time.After(100 * time.Millisecond)
		}
	}
}

func copyMessage(msg Message, registeredOrders [NumFloors][NumButtons - 1]AckList, elevator, floor, id int, btn Button) [NumFloors][NumButtons - 1]AckList {
	registeredOrders[floor][btn].ImplicitAcks[id] = msg.RegisteredOrders[floor][btn].ImplicitAcks[elevator]
	registeredOrders[floor][btn].ImplicitAcks[elevator] = msg.RegisteredOrders[floor][btn].ImplicitAcks[elevator]
	registeredOrders[floor][btn].DesignatedElevator = msg.RegisteredOrders[floor][btn].DesignatedElevator
	return registeredOrders
}
