package main

import (
	. "Hardware"
	. "constants"
	"fmt"
	//. "math/rand"
	"time"
)

func main() {
	var dir Direction = 2
	fmt.Println(dir)
	e := ET_Simulation
	Init(e)
	flag := 0
	SetMotorDirection(DirUp)
	for {
		if GetFloorSensorSignal() == 3 && flag != 1 {
			SetMotorDirection(DirStop)
			SetDoorOpenLamp(1)
			time.Sleep(30 * time.Millisecond)
			SetDoorOpenLamp(0)
			SetMotorDirection(DirDown)
			flag = 1
		}
		if GetFloorSensorSignal() == 0 && flag != 0 {
			SetMotorDirection(DirStop)
			SetDoorOpenLamp(1)
			time.Sleep(30 * time.Millisecond)
			SetDoorOpenLamp(0)
			SetMotorDirection(DirUp)
			flag = 0
		}
	}
}
