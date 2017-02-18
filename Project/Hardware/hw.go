package Hardware

import (
	. "fmt"
	"net"
	"os"
	"sync"

	. "github.com/perkjelsvik/TTK4145-Sanntid/Project/constants"
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

// Elev_type type
type Elev_type int

const (
	ET_Comedi Elev_type = iota
	ET_Simulation
)

var elevatorType Elev_type = ET_Comedi
var conn *net.TCPConn
var mtx *sync.Mutex
var sim_port string = "15857"

func Init(e Elev_type) {
	elevatorType = e
	switch elevatorType {
	case ET_Comedi:
		initSuccess := io_init()

		if initSuccess == 0 {
			Println("Unable to initialize elevator hardware!")
			os.Exit(1)
		}

	case ET_Simulation:
		addr, err := net.ResolveTCPAddr("tcp4", ":"+sim_port)
		Println(err)
		conn, err = net.DialTCP("tcp4", nil, addr)
		Println(err)
		mtx = &sync.Mutex{}
	}

	for f := 0; f < NumFloors; f++ {
		for b := 0; b < NumButtons; b++ {
			SetButtonLamp(Button(b), f, 0)
		}
	}

	setStopLamp(0)
	SetDoorOpenLamp(0)
	setFloorIndicator(0)
}

func SetMotorDirection(dirn Direction) {
	switch elevatorType {
	case ET_Comedi:
		if dirn == 0 {
			io_writeAnalog(MOTOR, 0)
		} else if dirn > 0 {
			io_clearBit(MOTORDIR)
			io_writeAnalog(MOTOR, MOTOR_SPEED)
		} else if dirn < 0 {
			io_setBit(MOTORDIR)
			io_writeAnalog(MOTOR, MOTOR_SPEED)
		}
	case ET_Simulation:
		mtx.Lock()
		conn.Write([]byte{1, byte(dirn), 0, 0})
		mtx.Unlock()
	}
}

func SetButtonLamp(btn Button, floor int, value int) {
	switch elevatorType {
	case ET_Comedi:
		if value > 0 {
			io_setBit(lamp_channel_matrix[floor][btn])
		} else {
			io_clearBit(lamp_channel_matrix[floor][btn])
		}
	case ET_Simulation:
		mtx.Lock()
		conn.Write([]byte{1, byte(btn), byte(floor), byte(value)})
		mtx.Unlock()
	}
}

func setFloorIndicator(floor int) {
	switch elevatorType {
	case ET_Comedi:
		// Binary encoding. One light must always be on.
		if floor&0x02 != 0 {
			io_setBit(LIGHT_FLOOR_IND1)
		} else {
			io_clearBit(LIGHT_FLOOR_IND1)
		}

		if floor&0x01 != 0 {
			io_setBit(LIGHT_FLOOR_IND2)
		} else {
			io_clearBit(LIGHT_FLOOR_IND2)
		}
	case ET_Simulation:
		mtx.Lock()
		conn.Write([]byte{3, byte(floor), 0, 0})
		mtx.Unlock()

	}
}

// SetDoorOpenLamp comment
func SetDoorOpenLamp(value int) {
	switch elevatorType {
	case ET_Comedi:
		if value > 0 {
			io_setBit(LIGHT_DOOR_OPEN)
		} else {
			io_clearBit(LIGHT_DOOR_OPEN)
		}
	case ET_Simulation:
		mtx.Lock()
		conn.Write([]byte{4, byte(value), 0, 0})
		mtx.Unlock()

	}
}

func setStopLamp(value int) {
	switch elevatorType {
	case ET_Comedi:
		if value > 0 {
			io_setBit(LIGHT_STOP)
		} else {
			io_clearBit(LIGHT_STOP)
		}
	case ET_Simulation:
		mtx.Lock()
		conn.Write([]byte{5, byte(value), 0, 0})
		mtx.Unlock()

	}
}

func getButtonSignal(btn Button, floor int) int {
	switch elevatorType {
	case ET_Comedi:
		return io_readBit(button_channel_matrix[floor][btn])
	case ET_Simulation:
		mtx.Lock()
		conn.Write([]byte{6, byte(btn), byte(floor), 0})
		buf := make([]byte, 4)
		conn.Read(buf)
		mtx.Unlock()
		return int(buf[1])
	}
	return 0
}

func GetFloorSensorSignal() int {
	switch elevatorType {
	case ET_Comedi:
		if io_readBit(SENSOR_FLOOR1) != 0 {
			return 0
		} else if io_readBit(SENSOR_FLOOR2) != 0 {
			return 1
		} else if io_readBit(SENSOR_FLOOR3) != 0 {
			return 2
		} else if io_readBit(SENSOR_FLOOR4) != 0 {
			return 3
		} else {
			return -1
		}
	case ET_Simulation:
		mtx.Lock()
		conn.Write([]byte{7, 0, 0, 0})
		buf := make([]byte, 4)
		conn.Read(buf)
		mtx.Unlock()
		if buf[1] == 0 {
			return -1
		} else {
			return int(buf[2])
		}
	}
	return 0
}

func getStopSignal() int {
	switch elevatorType {
	case ET_Comedi:
		return io_readBit(STOP)
	case ET_Simulation:
		mtx.Lock()
		conn.Write([]byte{8, 0, 0, 0})
		buf := make([]byte, 4)
		conn.Read(buf)
		mtx.Unlock()
		return int(buf[1])

	}
	return 0
}

func getObstructionSignal() int {
	switch elevatorType {
	case ET_Comedi:
		return io_readBit(OBSTRUCTION)
	case ET_Simulation:
		mtx.Lock()
		conn.Write([]byte{9, 0, 0, 0})
		buf := make([]byte, 4)
		conn.Read(buf)
		mtx.Unlock()
		return int(buf[1])

	}
	return 0
}
