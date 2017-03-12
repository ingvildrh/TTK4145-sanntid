package syncElevators

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"time"
    "reflect"

	"github.com/copier"
	. "github.com/perkjelsvik/TTK4145-sanntid/project/constants"
	"github.com/perkjelsvik/TTK4145-sanntid/project/networkCommunication/network/peers"
)

type SyncChannels struct {
	UpdateGovernor chan map[int]Elev
	UpdateSync     chan Elev
	OrderUpdate    chan Keypress
	IncomingMsg    chan Message
	OutgoingMsg    chan Message
	broadcastTimer <-chan time.Time
	PeerUpdate     chan peers.PeerUpdate
	PeerTxEnable   chan bool
}



/*

    Have assumed that:
        governor can receive map[int]*Elev, instead of [NumElevators]Elev
            (copy/dup of this map should not be necessary as long as Elev.Queue is array (not slice), like it is now)
        governor deals with lights (iterates through all Elev.Queue's, sets lights based on this)
        governor sends our own elev's queue to fsm-thing (ie sync doesnt send single new orders to fsm)
        fsm deletes the order from it's Elev.Queue, then sends its Elev here (maybe via governor?) (ie doesnt send single "completed order" here
            If this is not true, fix that in {select-case newOrder => switch-case Done=true}, see commented-out code
        

*/


func SYNC_loop(ch SyncChannels, id int) { //, syncBtnLights chan [NumFloors][NumButtons]bool) {

    var elevators   map[int]*Elev
    var hallOrders  [NumFloors][NumButtons-1]OrderAckStatus
    var peers       []int

	ch.broadcastTimer = time.Tick(100 * time.Millisecond).C
	for {
		select {

        // New copy of our own elevator state
		case elev := <-ch.UpdateSync:
            if !reflect.DeepEqual(elev, elevators[id]) {
                elevators[id] = elev
                ch.UpdateGovernor <- elevators.Dup()
            }

        // New order, either new or done 
/* TODO: New and Done orders are different enough to be on two separate channels */
		case newOrder := <-ch.OrderUpdate:
			switch newOrder.Done
            case false:
                ackStatus := hallOrders[newOrder.Floor][newOrder.Btn].AckStatus
                if ackStatus == Finished || ackStatus == Undefined {
                    hallOrders[newOrder.Floor][newOrder.Btn].DesignatedElevator = newOrder.DesignatedElevator
                    hallOrders[newOrder.Floor][newOrder.Btn].AckStatus = NotAcked
                    hallOrders[newOrder.Floor][newOrder.Btn].Acks = []int{id}
                }
            case true:
                if len(peers) == 0 || (len(peers) == 1 && peers[0] == id) {
                    hallOrders[newOrder.Floor][newOrder.Btn].AckStatus = Undefined
                } else {
                    hallOrders[newOrder.Floor][newOrder.Btn].AckStatus = Finished
                }
                hallOrders[newOrder.Floor][newOrder.Btn].Acks = []int{}
/*
                elevators[newOrder.DesignatedElevator].Queue[newOrder.Floor][newOrder.Btn] = false
*/
            }

        // New elev+hallorders bcast from peer (or self)
		case inMsg := <-ch.IncomingMsg:

            // Update elevator
            if inMsg.ID != id {
                _, exists := elevators[inMsg.ID]
                if !exists || !reflect.DeepEqual(elevators[inMsg.ID], inMsg.Elevator) {
                    elevators[inMsg.ID] = inMsg.Elevator
                    ch.UpdateGovernor <- elevators.Dup()
                }
            }

            // Update hall orders
            for f := range hallOrders {
                for b := range hallOrders[f] {

                    whenAcked = func(){
                        designated := inMsg.HallOrders[f][b].DesignatedElevator 
                        fmt.Printf("Hall order {floor:%v, button:%v} acknowledged. Assigned to: %v", f, b, designated)
                        if _, exists := elevators[designated]; exists {
                            elevators[designated].Queue[f][b] = true
                        } else {
                            // Can only happen during init, if an order is already assigned to an elevator we haven't heard from yet
                            // Could ignore this, but better to just take it ourselves, to be safe
                            elevators[id].Queue[f][b] = true
                        }
                        ch.UpdateGovernor <- elevators.Dup()                        
                    }
                    whenFinished = func(){
                        fmt.Printf("Hall order {floor:%v, button:%v} finished", f, b)
                        ch.UpdateGovernor <- elevators.Dup()      

                    }

                    switch hallOrders[f][b].AckStatus {
                    case Undefined:
                        // Assume remote state (as long as it is also not undefined)
                        if inMsg.ID != id {
                            switch inMsg.HallOrders[f][b].AckStatus {
                            case Finished:
                                hallOrders[f][b].DesignatedElevator = 0
                                hallOrders[f][b].AckStatus = Finished
                                hallOrders[f][b].Acks = []int{}
                            case NotAcked:
                                hallOrders[f][b].DesignatedElevator = inMsg.HallOrders[f][b].DesignatedElevator
                                hallOrders[f][b].AckStatus = NotAcked
                                hallOrders[f][b].Acks = append(inMsg.HallOrders[f][b].Acks, id)
                            case Acked:
                                hallOrders[f][b].DesignatedElevator = inMsg.HallOrders[f][b].DesignatedElevator
                                hallOrders[f][b].AckStatus = Acked
                                hallOrders[f][b].Acks = append(inMsg.HallOrders[f][b].Acks, id)
                                whenAcked()
                            }
                        }

                    case Finished:
                        // Transition to NotAcked if remote has started ack procedure
                        switch inMsg.HallOrders[f][b].AckStatus {
                        case NotAcked:
                            hallOrders[f][b].DesignatedElevator = inMsg.HallOrders[f][b].DesignatedElevator
                            hallOrders[f][b].AckStatus = NotAcked
                            hallOrders[f][b].Acks = append(inMsg.HallOrders[f][b].Acks, id)
                        }

                    case NotAcked:
                        // Transition to Acked if a) we see that all have acked, or b) remote says all have acked
                        switch inMsg.HallOrders[f][b].AckStatus {
                        case NotAcked:
                            hallOrders[f][b].Acks = append(append(hallOrders[f][b].Acks, inMsg.HallOrders[f][b].Acks...), id)
                            if containsAll(hallOrders[f][b].Acks, inMsg.HallOrders[f][b].Acks) {
                                hallOrders[f][b].AckStatus = Acked
                                whenAcked()
                            }
                        case Acked:
                            hallOrders[f][b].AckStatus = Acked
                            hallOrders[f][b].Acks = append(append(hallOrders[f][b].Acks, inMsg.HallOrders[f][b].Acks...), id)
                            whenAcked()
                        }

                    case Acked:
                        // Transition to Finished if remote has finished, or add any remaining acks (from late new peers)
                        switch inMsg.HallOrders[f][b].AckStatus {
                        case Finished:
                            hallOrders[f][b].AckStatus = Finished
                            hallOrders[f][b].Acks = []int{}
                            whenFinished()                            
                        case Acked:
                            hallOrders[f][b].Acks = append(append(hallOrders[f][b].Acks, inMsg.HallOrders[f][b].Acks...), id)
                        }
                        
                    }
                    
                    sort.Ints(hallOrders[f][b].Acks)
                    hallOrders[f][b].Acks = unique(hallOrders[f][b].Acks)
                    
                }
            }



        // New broadcast tick: send elev+hallorders
		case <-ch.broadcastTimer:
            ch.OutgoingMsg <- Message{id, elevators[id], hallOrders.Dup()}

		case p := <-ch.PeerUpdate:
            // Convert to []int
            peers = []int{}
            for i := range p.Peers {
                j, err := strconv.Atoi(p.Peers[i])
                if err != nil {
                    panic(err)
                }
                peers = append(peers, j)
            }
            
            // If we are alone on the network, we must be ready to accept "existing" orders from the rest of the network
            if len(peers) == 0 || (len(peers) == 1 && peers[0] == id) {
                for f := range hallOrders {
                    for b := range hallOrders[f] {
                        if hallOrders[f][b].AckStatus == Finished {
                            hallOrders[f][b].AckStatus == Undefined
                        }
                    }
                }
            }
		}
	}
}



func containsAll(a []int, b[]int) bool {
    for i := range a {
        if !contains(b, a[i]) {
            return false
        }
    }
    return true
}

func contains(a []int, b int) bool {
    for _, v := range a {
        if v == b {
            return true
        }
    }
    return false
}

func unique(a []int) []int {
    encountered := map[int]bool{}
    result := []int{}

    for v := range a {
        if !encountered[a[v]]
            encountered[a[v]] = true
            result = append(result, elements[v])
        }
    }
    return result
}



