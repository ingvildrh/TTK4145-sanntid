package syncElevators

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/copier"
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
			{- - - -} 		   {assignedID elev1 elev2 elev3}
{assignedID [{elev1, ack} {elev2, ack} {elev3, ack}]} {assignedID elev1 elev2 elev3}
{assignedID elev1 elev2 elev3} {assignedID elev1 elev2 elev3}
{assignedID elev1 elev2 elev3} 			 {- - - -}
*/

func SYNC_loop(ch SyncChannels, id int) { //, syncBtnLights chan [NumFloors][NumButtons]bool) {
	var (
		registeredOrders [NumFloors][NumButtons - 1]AckList
		elevList         [NumElevators]Elev
		sendMsg          Message
		//allAcked           []PeerElevator
		//allFinished        []PeerElevator
		numOnlineElevators int
		someUpdate         bool
	)
	// NOTE: status {0 0 0} trumps {-1 -1 -1}
	// NOTE: status {0 0 1} trumps {-1 -1 1 }
	// NOTE: possible to go from {0 0 0} -> {1 1 1} -> {-1 -1 -1} -> {0 0 0}
	// NOTE: allAcked := [NumElevatorsOnline]Acknowledge{Acked, Acked, Acked}

	// A quick fix to keep the local internal orders active after an elevator-reset.
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	select {
	case initMsg := <-ch.IncomingMsg:
		/*
			elevList[id] = initMsg.Elevator[id]
			copier.Copy(&registeredOrders, &initMsg.RegisteredOrders)*/
		ourPeer := PeerElevator{ID: id, Status: Undefined}
		for floor := 0; floor < NumFloors; floor++ {
			for btn := BtnUp; btn < BtnInside; btn++ {
				// QUESTION: Better way to do this?
				fmt.Println("Linje 66: ", initMsg.RegisteredOrders[floor][btn].ImplicitAcks)
				// FIXME: rename this variable please
				registeredOrders[floor][btn].ImplicitAcks = append(registeredOrders[floor][btn].ImplicitAcks, ourPeer)
				sort.Slice(registeredOrders[floor][btn].ImplicitAcks, func(i, j int) bool {
					return registeredOrders[floor][btn].ImplicitAcks[i].ID < registeredOrders[floor][btn].ImplicitAcks[j].ID
				})
			}
		}

	case <-timeout:
		fmt.Println("sync init timeout")
		ourPeer := PeerElevator{ID: id, Status: NotAcked}
		for floor := 0; floor < NumFloors; floor++ {
			for btn := BtnUp; btn < BtnInside; btn++ {
				registeredOrders[floor][btn].ImplicitAcks = append(registeredOrders[floor][btn].ImplicitAcks, ourPeer)
			}
		}
		break
	}
	fmt.Println(registeredOrders)
	// NOTE: burde vi importere constants som def eller liknende? mer lesbart
	ch.broadcastTimer = time.After(100 * time.Millisecond)
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
					registeredOrders[newOrder.Floor][newOrder.Btn].ImplicitAcks[id].Status = Finished
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
					registeredOrders[newOrder.Floor][newOrder.Btn].ImplicitAcks[id].Status = Acked
					fmt.Println("We Acked new order", PrintBtn(newOrder.Btn), "at floor", newOrder.Floor+1)
				}
			}

		case inMsg := <-ch.IncomingMsg:
			var msg Message
			//copier.Copy(&msg, &inMsg)
			//fmt.Println("adress: ", msg, inMsg)
			//os.Exit(1)
			if inMsg.ID == id {
				continue
			} else {
				//fmt.Println(time.Now())
				//fmt.Println("We received ", msg.RegisteredOrders)
				//someUpdate = false
				if inMsg.Elevator != elevList {
					fmt.Println("IN", inMsg.Elevator)
					fmt.Println()
					fmt.Println("CURR", elevList)
					os.Exit(1)
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
				for elevator := 0; elevator < numOnlineElevators; elevator++ {
					if elevator == id {
						continue
					}
					for floor := 0; floor < NumFloors; floor++ {
						for btn := BtnUp; btn < BtnInside; btn++ {
							msg := make([]PeerElevator, len(inMsg.RegisteredOrders[floor][btn].ImplicitAcks))
							copy(msg, inMsg.RegisteredOrders[floor][btn].ImplicitAcks)
							// IDEA: Could compress by having if new is +1 or -2 of our own status -> copy
							//fmt.Println("148", msg)
							//fmt.Println("elev:", elevator)
							switch msg[elevator].Status {
							case NotAcked:
								if registeredOrders[floor][btn].ImplicitAcks[id].Status == Finished {
									//registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
									registeredOrders[floor][btn].ImplicitAcks[id].Status = NotAcked
									registeredOrders[floor][btn].ImplicitAcks[elevator].Status = NotAcked
									//QUESTION: Shouldn't these be merged to one if-statement?
								} else if registeredOrders[floor][btn].ImplicitAcks[id].Status == Undefined {
									fmt.Println(registeredOrders[floor][btn].ImplicitAcks)
									//registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
									registeredOrders[floor][btn].ImplicitAcks[id].Status = NotAcked
									registeredOrders[floor][btn].ImplicitAcks[elevator].Status = NotAcked
								} else {
									registeredOrders[floor][btn].ImplicitAcks[elevator].Status = NotAcked
								}
							case Acked:
								if registeredOrders[floor][btn].ImplicitAcks[id].Status == NotAcked {
									fmt.Println("Order ", PrintBtn(btn), "in floor", floor+1, "has been acked!")
									//registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
									registeredOrders[floor][btn].ImplicitAcks[id].Status = Acked
									registeredOrders[floor][btn].ImplicitAcks[elevator].Status = Acked
								} else if registeredOrders[floor][btn].ImplicitAcks[id].Status == Undefined {
									//registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
									registeredOrders[floor][btn].ImplicitAcks[id].Status = Acked
									registeredOrders[floor][btn].ImplicitAcks[elevator].Status = Acked
								} else {
									registeredOrders[floor][btn].ImplicitAcks[elevator].Status = Acked
								}

								if allEquals(Acked, registeredOrders[floor][btn].ImplicitAcks) &&
									!elevList[id].Queue[floor][btn] &&
									registeredOrders[floor][btn].DesignatedElevator == id {
									fmt.Println("We've been assigned a new order!")
									elevList[id].Queue[floor][btn] = true
									someUpdate = true
								}
							case Finished:
								if registeredOrders[floor][btn].ImplicitAcks[id].Status == Acked {
									fmt.Println("Order ", PrintBtn(btn), "in floor", floor+1, "has been finished")
									fmt.Println("msg: ", msg[floor])
									fmt.Println("our: ", registeredOrders[floor])
									//registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
									registeredOrders[floor][btn].ImplicitAcks[id].Status = Finished
									registeredOrders[floor][btn].ImplicitAcks[elevator].Status = Finished
									fmt.Println("our: ", registeredOrders[floor])
								} else if registeredOrders[floor][btn].ImplicitAcks[id].Status == Undefined {
									//registeredOrders = copyMessage(msg, registeredOrders, elevator, floor, id, btn)
									registeredOrders[floor][btn].ImplicitAcks[id].Status = Finished
									registeredOrders[floor][btn].ImplicitAcks[elevator].Status = Finished
								} else {
									registeredOrders[floor][btn].ImplicitAcks[elevator].Status = Finished
								}

								if allEquals(Finished, registeredOrders[floor][btn].ImplicitAcks) {
									registeredOrders[floor][btn].ImplicitAcks[id].Status = NotAcked
									fmt.Println("All has acked Finished! NotAcking my Finished")
								}
							case Undefined:
								registeredOrders[floor][btn].ImplicitAcks[id].Status = NotAcked
							}
						}
					}
				}
				if someUpdate {
					ch.UpdateGovernor <- elevList
					someUpdate = false
				}
			}

		case <-ch.broadcastTimer:
			//fmt.Println("Hello to you")
			// NB: Don't know if this works AT ALL
			copier.Copy(&sendMsg.RegisteredOrders, &registeredOrders)
			//fmt.Println(registeredOrders)
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
			numOnlineElevators = len(p.Peers)
			if len(p.New) != 0 {
				newID, _ := strconv.Atoi(p.New)
				if newID == id {
					break
				}
				for floor := 0; floor < NumFloors; floor++ {
					for btn := BtnUp; btn < BtnInside; btn++ {
						fmt.Print("Linje 219:", registeredOrders[floor][btn])
						registeredOrders[floor][btn].ImplicitAcks = append(registeredOrders[floor][btn].ImplicitAcks,
							PeerElevator{ID: newID, Status: Undefined})
						sort.Slice(registeredOrders[floor][btn].ImplicitAcks, func(i, j int) bool {
							return registeredOrders[floor][btn].ImplicitAcks[i].ID < registeredOrders[floor][btn].ImplicitAcks[j].ID
						})
					}
				}
			}
			if len(p.Lost) != 0 {
				lostID, _ := strconv.Atoi(p.Lost[0])
				for floor := 0; floor < NumFloors; floor++ {
					for btn := BtnUp; btn < BtnInside; btn++ {
						if lostID == len(registeredOrders[floor][btn].ImplicitAcks) {
							registeredOrders[floor][btn].ImplicitAcks = append(registeredOrders[floor][btn].ImplicitAcks[:lostID],
								registeredOrders[floor][btn].ImplicitAcks[:lostID-1]...)
						} else {
							registeredOrders[floor][btn].ImplicitAcks = append(registeredOrders[floor][btn].ImplicitAcks[:lostID],
								registeredOrders[floor][btn].ImplicitAcks[lostID+1:]...)
						}
					}
				}
			}
		}
	}
}

// FIXME: Change name to copyAckList? copyAckStatus? or something else?
func copyMessage(msg Message, registeredOrders [NumFloors][NumButtons - 1]AckList, elevator, floor, id int, btn Button) [NumFloors][NumButtons - 1]AckList {
	tmpPeerList := make([]PeerElevator, len(msg.RegisteredOrders[floor][btn].ImplicitAcks))
	copy(tmpPeerList, msg.RegisteredOrders[floor][btn].ImplicitAcks)
	fmt.Println("tmp", tmpPeerList)
	registeredOrders[floor][btn].ImplicitAcks[id] = msg.RegisteredOrders[floor][btn].ImplicitAcks[elevator]
	registeredOrders[floor][btn].ImplicitAcks[elevator] = msg.RegisteredOrders[floor][btn].ImplicitAcks[elevator]
	registeredOrders[floor][btn].DesignatedElevator = msg.RegisteredOrders[floor][btn].DesignatedElevator
	return registeredOrders
}

func allEquals(status Acknowledge, v ...interface{}) bool {
	if v[0] != status {
		return false
	}
	if len(v) < 2 {
		return true
	}
	return reflect.DeepEqual(v[:len(v)-1], v[1:])
}
