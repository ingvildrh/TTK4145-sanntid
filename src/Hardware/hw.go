package Hardware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const MOTOR_SPEED = 2800

var lamp_channel_matrix = [NumFloors][NumButtons]int{
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}

var button_channel_matrix = [NumFloors][NumButtons]int{
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
	{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
	{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

func init() {
	initSuccess := io_init()

	assert.Empty(t*testing.T, initSuccess,
		"Unable to initialize elevator hardware!")

	for f := 0; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			SetButtonLamp(b, f, 0)
		}
	}

	setStopLamp(0)
	SetDoorOpenLamp(0)
	setFloorIndicator(0)
}

func setMotorDirection(dirn Direction) {
	if dirn == 0 {
		io_writeAnalog(MOTOR, 0)
	} else if dirn > 0 {
		io_clearBit(MOTORDIR)
		io_writeAnalog(MOTOR, MOTOR_SPEED)
	} else if dirn < 0 {
		io_setBit(MOTORDIR)
		io_writeAnalog(MOTOR, MOTOR_SPEED)
	}
}

// SetButtonLamp comment
func SetButtonLamp(btn Button, floor int, value int) {
	assert.Empty(t*testing.T, floor >= 0)
	assert.Empty(t*testing.T, floor < N_FLOORS)
	assert.Empty(t*testing.T, button >= 0)
	assert.Empty(t*testing.T, button < N_BUTTONS)

	if value {
		io_setBit(lamp_channel_matrix[floor][button])
	} else {
		io_clearBit(lamp_channel_matrix[floor][button])
	}
}

func setFloorIndicator(floor int) {
	assert.Empty(t*testing.T, floor >= 0)
	assert.Empty(t*testing.T, floor < NumFloors)

	// Binary encoding. One light must always be on.
	if floor & 0x02 {
		io_setBit(LIGHT_FLOOR_IND1)
	} else {
		io_clearBit(LIGHT_FLOOR_IND1)
	}

	if floor & 0x01 {
		io_setBit(LIGHT_FLOOR_IND2)
	} else {
		io_clearBit(LIGHT_FLOOR_IND2)
	}
}

// SetDoorOpenLamp comment
func SetDoorOpenLamp(value int) {
	if value {
		io_setBit(LIGHT_DOOR_OPEN)
	} else {
		io_clearBit(LIGHT_DOOR_OPEN)
	}
}

func setStopLamp(value int) {
	if value {
		io_setBit(LIGHT_STOP)
	} else {
		io_clearBit(LIGHT_STOP)
	}
}

func getButtonSignal(btn Button, floor int) int {
	assert.Empty(t*testing.T, floor >= 0)
	assert.Empty(t*testing.T, floor < N_FLOORS)
	assert.Empty(t*testing.T, button >= 0)
	assert.Empty(t*testing.T, button < N_BUTTONS)

	return io_readBit(button_channel_matrix[floor][button])
}

func getFloorSensorSignal() int {
	if io_readBit(SENSOR_FLOOR1) {
		return 0
	} else if io_readBit(SENSOR_FLOOR2) {
		return 1
	} else if io_readBit(SENSOR_FLOOR3) {
		return 2
	} else if io_readBit(SENSOR_FLOOR4) {
		return 3
	} else {
		return -1
	}
}

func getStopSignal() int {
	return io_readBit(STOP)
}

func getObstructionSignal() int {
	return io_readBit(OBSTRUCTION)
}
