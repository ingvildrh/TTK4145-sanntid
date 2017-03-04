package constants

// Scaleable declaration of #floors and #elevators
const (
	NumFloors    = 4
	NumElevators = 3
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
}

type Elev struct {
	State int
	Dir   Direction
	Floor int
	Queue [NumFloors][NumButtons]bool
}

type AckMatrix struct {
	OrderUp            bool
	OrderDown          bool
	DesignatedElevator int
	ImplicitAcks       [NumElevators]Acknowledge
}

type Msg struct {
	Elevator         [NumElevators]Elev
	RegisteredOrders [NumFloors]AckMatrix
}

type Acknowledge int

const (
	Finished Acknowledge = iota - 1
	NotAcked
	Acked
)
