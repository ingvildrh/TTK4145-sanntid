package syncElevators

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	"github.com/perkjelsvik/TTK4145-sanntid/project/networkCommunication/network/peers"
)

type SyncChannels struct {
	UpdateGovernor  chan [NumElevators]Elev
	UpdateSync      chan Elev
	OrderUpdate     chan Keypress
	IncomingMsg     chan Message
	OutgoingMsg     chan Message
	PeerUpdate      chan peers.PeerUpdate
	PeerTxEnable    chan bool
	OnlineElevators chan [NumElevators]bool
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
		onlineList       [NumElevators]bool
		recentlyDied     [NumElevators]bool
		someUpdate       bool
		offline          bool
		singleMode       bool
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
		// NB: Did we mean to just copy our own information and not everything else?
		elevList = initMsg.Elevator
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
		offline = true
		break
	}
	// NOTE: burde vi importere constants som def eller liknende? mer lesbart
	broadcastTicker := time.NewTicker(100 * time.Millisecond)
	reassignTimer := time.NewTimer(5 * time.Second)
	singleModeTicker := time.NewTicker(100 * time.Millisecond) //for testing purposes
	singleModeTicker.Stop()
	reassignTimer.Stop()
	/*simTimer := time.NewTimer(20 * time.Second)
	simTimer.Stop()
	if id != 2 {
		simTimer = time.NewTimer(20 * time.Second) //for testing purposes
	}
	*/

	for {

		if offline {
			if onlineList[id] {
				offline = false
				reInitTimer := time.NewTimer(1000 * time.Millisecond)
			REINIT:
				for {
					select {
					case reInitMsg := <-ch.IncomingMsg:
						if reInitMsg.Elevator != elevList && reInitMsg.ID != id {
							tmpElevator := elevList[id]
							elevList = reInitMsg.Elevator
							elevList[id] = tmpElevator
							someUpdate = true
							reInitTimer.Stop()
							break REINIT
						}
					case <-reInitTimer.C:
						break REINIT
					}
				}
			}
		}

		if lostID != -1 {
			fmt.Println("ELEVATOR", lostID, "DIED")
			recentlyDied[lostID] = true
			lostID = -1
		}

		if isSingleElevator(onlineList, id) {
			if !singleMode {
				singleMode = true
				singleModeTicker = time.NewTicker(100 * time.Millisecond)
			}
		}

		select {
		/*
			case <-simTimer.C:
					if !offline {
						offline = true
						ch.PeerTxEnable <- false
						onlineList = [NumElevators]bool{}
						fmt.Println("------------- LOST INTERNET -------------")
						simTimer.Reset(20 * time.Second)
					} else {
						onlineList[id] = true
						ch.PeerTxEnable <- true
						fmt.Println("------------- REGAINED INTERNET -------------")
						simTimer.Reset(20 * time.Second)
					}
		*/

		case newElev := <-ch.UpdateSync:
			oldQueue := elevList[id].Queue

			// FIXME: We REALLY have to move state definitions to constants
			if newElev.State == -1 {
				ch.PeerTxEnable <- false
			} else if newElev.State != -1 && elevList[id].State == -1 {
				ch.PeerTxEnable <- true
			}

			elevList[id] = newElev
			elevList[id].Queue = oldQueue
			someUpdate = true

		case newOrder := <-ch.OrderUpdate:
			if newOrder.Done {
				// NB: Here we clear all orders from floor
				elevList[id].Queue[newOrder.Floor] = [NumButtons]bool{}
				someUpdate = true
				if newOrder.Btn != BtnInside {
					// FIXME: this is to prevent out of index because of BtnInside. Need better fix.
					registeredOrders[newOrder.Floor][BtnUp].ImplicitAcks[id] = Finished
					registeredOrders[newOrder.Floor][BtnDown].ImplicitAcks[id] = Finished
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
			if msg.ID == id || !onlineList[msg.ID] || !onlineList[id] {
				continue
			} else {
				// NB: need to handle the case where an elevator has lost motor, but has net (peer disabled, network not disabled)
				if msg.Elevator != elevList {
					tmpElevator := elevList[id]
					elevList = msg.Elevator
					elevList[id] = tmpElevator
					someUpdate = true
				}
				//fmt.Println("Hello from me")
				// IDEA: Have another ack-state ackButNotAllAcked.
				for elevator := 0; elevator < NumElevators; elevator++ {
					if elevator == id || !onlineList[msg.ID] || !onlineList[id] {
						continue
					}
					for floor := 0; floor < NumFloors; floor++ {
						for btn := BtnUp; btn < BtnInside; btn++ {
							// IDEA: Could compress by having if new is +1 or -2 of our own status -> copy
							switch msg.RegisteredOrders[floor][btn].ImplicitAcks[elevator] {

							case NotAcked:
								if registeredOrders[floor][btn].ImplicitAcks[id] == Finished {
									registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
								} else if registeredOrders[floor][btn].ImplicitAcks[elevator] != NotAcked {
									registeredOrders[floor][btn].ImplicitAcks[elevator] = NotAcked
								}

							case Acked:
								if registeredOrders[floor][btn].ImplicitAcks[id] == NotAcked {
									fmt.Println("Order ", PrintBtn(btn), "from ", msg.ID, "in floor", floor+1, "has been acked!")
									registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
								} else if registeredOrders[floor][btn].ImplicitAcks[elevator] != Acked {
									registeredOrders[floor][btn].ImplicitAcks[elevator] = Acked
								}
								if checkAllAckStatus(onlineList, registeredOrders[floor][btn].ImplicitAcks, Acked) &&
									!elevList[id].Queue[floor][btn] &&
									registeredOrders[floor][btn].DesignatedElevator == id {
									fmt.Println("We've been assigned a new order!")
									elevList[id].Queue[floor][btn] = true
									someUpdate = true
								}

							case Finished:
								if registeredOrders[floor][btn].ImplicitAcks[id] == Acked {
									//fmt.Println("Order ", PrintBtn(btn), "in floor", floor+1, "has been finished")
									//fmt.Println("msg: ", msg.RegisteredOrders[floor])
									//fmt.Println("our: ", registeredOrders[floor])
									registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
									//fmt.Println("our: ", registeredOrders[floor])
								} else if registeredOrders[floor][btn].ImplicitAcks[elevator] != Finished {
									registeredOrders[floor][btn].ImplicitAcks[elevator] = Finished
								}

								if checkAllAckStatus(onlineList, registeredOrders[floor][btn].ImplicitAcks, Finished) {
									registeredOrders[floor][btn].ImplicitAcks[id] = NotAcked
									if registeredOrders[floor][btn].DesignatedElevator == id {
										elevList[id].Queue[floor][btn] = false
										someUpdate = true
									}
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
			}

		case <-singleModeTicker.C:
			// FIXME: Not properly checked, also this print is invalid
			//fmt.Println("offlineTick")
			for floor := 0; floor < NumFloors; floor++ {
				for btn := BtnUp; btn < BtnInside; btn++ {
					if registeredOrders[floor][btn].ImplicitAcks[id] == Acked &&
						!elevList[id].Queue[floor][btn] {
						fmt.Println("We've been assigned a new order!")
						elevList[id].Queue[floor][btn] = true
						someUpdate = true
					}
					if registeredOrders[floor][btn].ImplicitAcks[id] == Finished {
						registeredOrders[floor][btn].ImplicitAcks[id] = NotAcked
					}

				}
			}
			if someUpdate {
				ch.UpdateGovernor <- elevList
				someUpdate = false
			}

		case <-broadcastTicker.C:
			//fmt.Println("Hello to you")
			sendMsg.RegisteredOrders = registeredOrders
			sendMsg.Elevator = elevList
			sendMsg.ID = id
			if !offline {
				ch.OutgoingMsg <- sendMsg
			}
			//fmt.Println("We sent", sendMsg.RegisteredOrders)

		case p := <-ch.PeerUpdate:
			// FIXME: Need a zeroStatus (bool) to handle One Elevator Alive and regaining connection
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			// NB: We are dependent on string ID to be correct (0, 1 or 2)
			if len(p.Peers) == 0 {
				offline = true
			} else if len(p.Peers) == 1 {
				singleMode = true
			}
			if len(p.New) > 0 {
				newID, _ := strconv.Atoi(p.New)
				onlineList[newID] = true
			} else if len(p.Lost) > 0 {
				lostID, _ = strconv.Atoi(p.Lost[0])
				onlineList[lostID] = false
				if elevList[lostID].Queue != [NumFloors][NumButtons]bool{} && !recentlyDied[lostID] {
					reassignTimer.Reset(1 * time.Second)
				}
			}
			fmt.Println("online changed: ", onlineList)
			tmpList := onlineList
			// NB: is this safe? if not, do as goroutine!
			go func() { ch.OnlineElevators <- tmpList }()

		case <-reassignTimer.C:
			for elevator := 0; elevator < NumElevators; elevator++ {
				if !recentlyDied[elevator] {
					continue
				}
				recentlyDied[elevator] = false
				for floor := 0; floor < NumFloors; floor++ {
					for btn := BtnUp; btn < BtnInside; btn++ {
						if elevList[elevator].Queue[floor][btn] {
							elevList[id].Queue[floor][btn] = true
							elevList[elevator].Queue[floor][btn] = false
						}
					}
				}
			}
			ch.UpdateGovernor <- elevList
			someUpdate = false
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

func checkAllAckStatus(onlineList [NumElevators]bool, ImplicitAcks [NumElevators]Acknowledge, status Acknowledge) bool {
	for elev := 0; elev < NumElevators; elev++ {
		if !onlineList[elev] {
			continue
		}
		if ImplicitAcks[elev] != status {
			return false
		}
	}
	return true
}

func isSingleElevator(onlineList [NumElevators]bool, id int) bool {
	for elev := 0; elev < NumElevators; elev++ {
		if onlineList[elev] && elev != id {
			return false
		}
	}
	return true
}
