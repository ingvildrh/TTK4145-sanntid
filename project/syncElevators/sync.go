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
	designatedElevator := id
	// NOTE: burde vi importere constants som def eller liknende? mer lesbart
	for {
		select {
		case tmpElev := <-ch.UpdateSync:
			tmpQueue := elevList[id].Queue
			elevList[id] = tmpElev
			elevList[id].Queue = tmpQueue

		case newOrder := <-ch.OrderUpdate:
			if newOrder.Done {
				// NB: Here we clear all orders from floor
				elevList[id].Queue[newOrder.Floor] = [NumButtons]bool{}
				if newOrder.Btn != BtnInside {
					// FIXME: this is to prevent out of index because of BtnInside. Need better fix.
					registeredOrders[newOrder.Floor][newOrder.Btn].ImplicitAcks[id] = Finished
				}
			} else {
				if newOrder.Btn == BtnInside {
					// NB: Should probably send on net before adding to the queue. Exactly how unclear for now. To avoid immediate death after internal light on
					elevList[id].Queue[newOrder.Floor][newOrder.Btn] = true
				} else {
					registeredOrders[newOrder.Floor][newOrder.Btn].DesignatedElevator = newOrder.DesignatedElevator
					//NB: this is for testing purposes
					registeredOrders[newOrder.Floor][newOrder.Btn].ImplicitAcks[id] = Acked
				}
				// NB: This seems like a bad idea, bound to be Deadlock
				// // sende intern knappebestilling tilbake!!
				// ch.UpdateGovernor <- elevList
			}

		case msg := <-ch.IncomingMsg:
			someUpdate := false
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
							}

						case Acked:
							if registeredOrders[floor][btn].ImplicitAcks[id] == NotAcked {
								registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
							}
							if registeredOrders[floor][btn].ImplicitAcks == allAcked &&
								!elevList[designatedElevator].Queue[floor][btn] {
								designatedElevator = registeredOrders[floor][btn].DesignatedElevator
								elevList[designatedElevator].Queue[floor][btn] = true
								someUpdate = true
							}
						case Finished:
							if registeredOrders[floor][btn].ImplicitAcks[id] == Acked {
								registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
							}
						}
					}
				}
			}
			if someUpdate {
				ch.UpdateGovernor <- elevList
			}
		//FIXME: Should probably move these to outoing thread
		//QUESTION: How to share elevList between the threads in the best way?

		case <-ch.broadcastTimer:
			//fmt.Println("Hello to you")
			sendMsg.RegisteredOrders = registeredOrders
			sendMsg.Elevator = elevList
			ch.OutgoingMsg <- sendMsg
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
