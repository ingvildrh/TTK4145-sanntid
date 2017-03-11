package constants

import "os/exec"

// Scaleable declaration of #floors and #elevators
const (
	NumFloors = 4
	//FIXME: Change back to 3 elevators
	NumElevators = 2
	NumButtons   = 3
)

// Direction type
type Direction int

// Motor Directions
const (
	DirDown Direction = iota - 1
	DirStop
	DirUp
)

// Button type
type Button int

// Button mapping
const (
	BtnUp Button = iota
	BtnDown
	BtnInside
)

// Keypress struct for button type and floor location
type Keypress struct {
	Floor              int
	Btn                Button
	DesignatedElevator int
	Done               bool
}

type Elev struct {
	State int
	Dir   Direction
	Floor int
	//IDEA: make Queue private (queue)
	Queue [NumFloors][NumButtons]bool
}

type PeerElevator struct {
	ID     int
	Status Acknowledge
}

type AckList struct {
	DesignatedElevator int
	//IDEA: Make this dynamic after online/offline elevators
	ImplicitAcks []PeerElevator
}

type Message struct {
	Elevator         [NumElevators]Elev
	RegisteredOrders [NumFloors][NumButtons - 1]AckList
	ID               int
}

type Acknowledge int

const (
	Finished Acknowledge = iota - 1
	Undefined
	NotAcked
	Acked
)

// PrintBtn NB: er kun hjelpefunksjon, burde fjernes!
func PrintBtn(btn Button) string {
	switch btn {
	case BtnUp:
		return "UP"
	case BtnDown:
		return "DOWN"
	case BtnInside:
		return "INSIDE"
	default:
		return "ERROR"
	}
}

// Restart command used to spawn a new elevator process in case of power loss (ctrl+c)
var Restart = exec.Command("gnome-terminal", "-x", "sh", "-c", "sleep 3s;  go run main.go -simPort=9997 -ID=0 -run=sim")
