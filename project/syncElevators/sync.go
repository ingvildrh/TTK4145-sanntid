package syncElevators

import (
	"fmt"
	"time"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	"github.com/perkjelsvik/TTK4145-sanntid/project/networkCommunication/network/peers"
)

type SyncChannels struct {
	UpdateGovernor chan [NumElevators]Elev
	UpdateSync     chan Elev
	OrderUpdate    chan Keypress
	IncomingMsg    chan Message
	OutgoingMsg    chan Message
	broadcastTimer <-chan time.Time
	PeerUpdate     chan peers.PeerUpdate
	PeerTxEnable   chan bool
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
	var registeredOrders [NumFloors][NumButtons - 1]AckList
	var elevList [NumElevators]Elev
	var sendMsg Message
	var allAcked [NumElevators]Acknowledge
	var allFinished [NumElevators]Acknowledge
	var allNotAcked [NumElevators]Acknowledge
	someUpdate := false
	// NOTE: status {0 0 0} trumps {-1 -1 -1}
	// NOTE: status {0 0 1} trumps {-1 -1 1 }
	// NOTE: possible to go from {0 0 0} -> {1 1 1} -> {-1 -1 -1} -> {0 0 0}
	// NOTE: allAcked := [NumElevators]Acknowledge{Acked, Acked, Acked}
	for i := 0; i < NumElevators; i++ {
		allAcked[i] = Acked
		allFinished[i] = Finished
		allNotAcked[i] = NotAcked
	}
	ch.broadcastTimer = time.After(100 * time.Millisecond)
	// NOTE: burde vi importere constants som def eller liknende? mer lesbart
	for {
		select {
		case tmpElev := <-ch.UpdateSync:
			tmpQueue := elevList[id].Queue
			elevList[id] = tmpElev
			elevList[id].Queue = tmpQueue
			someUpdate = true

		case newOrder := <-ch.OrderUpdate:
			if newOrder.Done {
				// NB: Here we clear all orders from floor
				elevList[id].Queue[newOrder.Floor] = [NumButtons]bool{}
				someUpdate = true
				if newOrder.Btn != BtnInside {
					// FIXME: this is to prevent out of index because of BtnInside. Need better fix.
					registeredOrders[newOrder.Floor][newOrder.Btn].ImplicitAcks[id] = Finished
					fmt.Println("We Finished order", PrintBtn(newOrder.Btn), "at floor", newOrder.Floor+1)
				}
			} else {
				if newOrder.Btn == BtnInside {
					// NB: Should probably send on net before adding to the queue. Exactly how unclear for now. To avoid immediate death after internal light on
					elevList[id].Queue[newOrder.Floor][newOrder.Btn] = true
					someUpdate = true
				} else {
					registeredOrders[newOrder.Floor][newOrder.Btn].DesignatedElevator = newOrder.DesignatedElevator
					//NB: this is for testing purposes
					registeredOrders[newOrder.Floor][newOrder.Btn].ImplicitAcks[id] = Acked
					fmt.Println("We Acked new order", PrintBtn(newOrder.Btn), "at floor", newOrder.Floor+1)
				}
				// NB: This seems like a bad idea, bound to be Deadlock
				// // sende intern knappebestilling tilbake!!
				// ch.UpdateGovernor <- elevList
			}

		case msg := <-ch.IncomingMsg:
			if msg.ID == id {

			} else {
				//fmt.Println(time.Now())
				//fmt.Println("We received ", msg.RegisteredOrders)
				//someUpdate = false
				if msg.Elevator != elevList {
					fmt.Println("FUNKER")
					tmpElevator := elevList[id]
					//fmt.Println("tmpQueue: ", tmpQueue)
					elevList = msg.Elevator
					//fmt.Println("elevList: ", elevList[id].Queue)
					elevList[id] = tmpElevator
					someUpdate = true
				}
				//fmt.Println("Hello from me")
				// IDEA: Have another ack-state ackButNotAllAcked.
				for elevator := 0; elevator < NumElevators; elevator++ {
					if elevator == id {
						continue
					}
					for floor := 0; floor < NumFloors; floor++ {
						for btn := BtnUp; btn < BtnInside; btn++ {
							// IDEA: Could compress by having if new is +1 or -2 of our own status -> copy
							switch msg.RegisteredOrders[floor][btn].ImplicitAcks[elevator] {
							case NotAcked:
								if registeredOrders[floor][btn].ImplicitAcks[id] == Finished {
									registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
								} else {
									registeredOrders[floor][btn].ImplicitAcks[elevator] = NotAcked
								}
							case Acked:
								if registeredOrders[floor][btn].ImplicitAcks[id] == NotAcked {
									fmt.Println("Order ", PrintBtn(btn), "in floor", floor+1, "has been acked!")
									registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
								} else {
									registeredOrders[floor][btn].ImplicitAcks[elevator] = Acked
								}
								if registeredOrders[floor][btn].ImplicitAcks == allAcked &&
									!elevList[id].Queue[floor][btn] &&
									registeredOrders[floor][btn].DesignatedElevator == id {
									fmt.Println("We've been assigned a new order!")
									elevList[id].Queue[floor][btn] = true
									someUpdate = true
								}

							case Finished:
								if registeredOrders[floor][btn].ImplicitAcks[id] == Acked {
									fmt.Println("Order ", PrintBtn(btn), "in floor", floor+1, "has been finished")
									fmt.Println("msg: ", msg.RegisteredOrders[floor])
									fmt.Println("our: ", registeredOrders[floor])
									registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
									fmt.Println("our: ", registeredOrders[floor])
								} else {
									registeredOrders[floor][btn].ImplicitAcks[elevator] = Finished
								}
								if registeredOrders[floor][btn].ImplicitAcks == allFinished {
									registeredOrders[floor][btn].ImplicitAcks[id] = NotAcked
									fmt.Println("All has acked Finished! NotAcking my Finished")
								}
							}
						}
					}
				}
				if someUpdate {
					ch.UpdateGovernor <- elevList
					someUpdate = false
				}
				//FIXME: Should probably move these to outoing thread
				//QUESTION: How to share elevList between the threads in the best way?
				//fmt.Println(time.Now())
			}
		case <-ch.broadcastTimer:
			//fmt.Println("Hello to you")
			sendMsg.RegisteredOrders = registeredOrders
			sendMsg.Elevator = elevList
			sendMsg.ID = id
			ch.OutgoingMsg <- sendMsg
			//fmt.Println("We sent", sendMsg.RegisteredOrders)
			ch.broadcastTimer = time.After(100 * time.Millisecond)

		case p := <-ch.PeerUpdate:
			// FIXME: Need a zeroStatus (bool) to handle One Elevator Alive and regaining connection
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
		}
	}
}

// FIXME: Change name to copyAckList? copyAckStatus? or something else?
func copyMessage(msg Message, registeredOrders [NumFloors][NumButtons - 1]AckList, elevator, floor, id int, btn Button) [NumFloors][NumButtons - 1]AckList {
	registeredOrders[floor][btn].ImplicitAcks[id] = msg.RegisteredOrders[floor][btn].ImplicitAcks[elevator]
	registeredOrders[floor][btn].ImplicitAcks[elevator] = msg.RegisteredOrders[floor][btn].ImplicitAcks[elevator]
	registeredOrders[floor][btn].DesignatedElevator = msg.RegisteredOrders[floor][btn].DesignatedElevator
	return registeredOrders
}

// Function to check for changes
// NB: Should be unneccessary
func newInformation(msg Message, elevList [NumElevators]Elev, elevator, floor int, btn Button) bool {
	if msg.Elevator[elevator].Queue[floor][btn] != elevList[elevator].Queue[floor][btn] ||
		msg.Elevator[elevator].Dir != elevList[elevator].Dir ||
		msg.Elevator[elevator].State != elevList[elevator].State ||
		msg.Elevator[elevator].Floor != elevList[elevator].Floor {
		return true
	}
	return false
}
