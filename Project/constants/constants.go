package Constants

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

//Button type
type Button int

// Button mapping
const (
	BtnUp Button = iota
	BtnDown
	BtnInside
)

// Keypress struct for button type and floor location
type Keypress struct {
	Button int
	Floor  int
}
