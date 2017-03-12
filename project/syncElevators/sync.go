package syncElevators

import (
	"fmt"
	"strconv"
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
	reassignTimer  <-chan time.Time
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

func SYNC_loop(ch SyncChannels, id int) { //, syncBtnLights chan [NumFloors][NumButtons]bool) {
	var (
		registeredOrders [NumFloors][NumButtons - 1]AckList
		elevList         [NumElevators]Elev
		sendMsg          Message
		onlineElevators  [NumElevators]bool
		recentlyDied     [NumElevators]bool
		someUpdate       bool
		lostID           int
	)
	// NOTE: status {0 0 0} trumps {-1 -1 -1}
	// NOTE: status {0 0 1} trumps {-1 -1 1 }
	// NOTE: possible to go from {0 0 0} -> {1 1 1} -> {-1 -1 -1} -> {0 0 0}
	// NOTE: allAcked := [NumElevators]Acknowledge{Acked, Acked, Acked}

	// A quick fix to keep the local internal orders active after an elevator-reset.
	timeout := make(chan bool, 1)
	lostID = -1
	go func() { time.Sleep(1 * time.Second); timeout <- true }()
	select {
	case initMsg := <-ch.IncomingMsg:
		elevList[id] = initMsg.Elevator[id]
		registeredOrders = initMsg.RegisteredOrders
		fmt.Println("---------------------------- INIT ----------------------------")
		fmt.Println()
		for f := 0; f < NumFloors; f++ {
			fmt.Println(elevList[id].Queue[f], "\t", registeredOrders[f])
		}
		fmt.Println()
		fmt.Println("---------------------------- INIT DONE ------------------------")
		someUpdate = true
	case <-timeout:
		break
	}
	// NOTE: burde vi importere constants som def eller liknende? mer lesbart
	ch.broadcastTimer = time.After(100 * time.Millisecond)
	for {
		if lostID != -1 {
			fmt.Println("ELEVATOR", lostID, "DIED")
			recentlyDied[lostID] = true
			lostID = -1
		}
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
			}

		case msg := <-ch.IncomingMsg:
			// TODO: Must be able to run if only one alive
			if msg.ID == id {

			} else {
				if msg.Elevator != elevList {
					fmt.Println("FUNKER")
					tmpElevator := elevList[id]
					elevList = msg.Elevator
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

								if checkAllAckStatus(onlineElevators, registeredOrders[floor][btn].ImplicitAcks, Acked) &&
									!elevList[id].Queue[floor][btn] &&
									registeredOrders[floor][btn].DesignatedElevator == id {
									fmt.Println("We've been assigned a new order!")
									elevList[id].Queue[floor][btn] = true
									someUpdate = true
								}
								ch.broadcastTimer = time.After(100 * time.Millisecond)
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

								if checkAllAckStatus(onlineElevators, registeredOrders[floor][btn].ImplicitAcks, Finished) {
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
			// NB: We are dependent on string ID to be correct (0, 1 or 2)
			if len(p.New) > 0 {
				newID, _ := strconv.Atoi(p.New)
				onlineElevators[newID] = true
			} else if len(p.Lost) > 0 {
				lostID, _ = strconv.Atoi(p.Lost[0])
				onlineElevators[lostID] = false
				if elevList[lostID].Queue != [NumFloors][NumButtons]bool{} && !recentlyDied[lostID] {
					ch.reassignTimer = time.After(5 * time.Second)
				}
			}
		case <-ch.reassignTimer:
			for elevator := 0; elevator < NumElevators; elevator++ {
				if recentlyDied[elevator] == false {
					continue
				}
				recentlyDied[elevator] = false
				for floor := 0; floor < NumFloors; floor++ {
					for btn := BtnUp; btn < BtnInside; btn++ {
						fmt.Println(floor, btn)
						if elevList[elevator].Queue[floor][btn] {
							for elev := 0; elev < NumElevators; elev++ {
								if onlineElevators[elev] == false {
									continue
								}
								elevList[elev].Queue[floor][btn] = true
								elevList[elevator].Queue[floor][btn] = false
								registeredOrders[floor][btn].DesignatedElevator = elev
								elev = NumElevators
							}
						}
					}
				}
			}
			ch.UpdateGovernor <- elevList
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

func checkAllAckStatus(onlineElevators [NumElevators]bool, ImplicitAcks [NumElevators]Acknowledge, status Acknowledge) bool {
	for elev := 0; elev < NumElevators; elev++ {
		if onlineElevators[elev] == false {
			continue
		}
		if ImplicitAcks[elev] != status {
			return false
		}
	}
	return true
}
