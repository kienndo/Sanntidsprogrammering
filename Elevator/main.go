package main

import (
	"Sanntidsprogrammering/Elevator/devices"
	"Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	timer "Sanntidsprogrammering/Elevator/timer"
	"time"
)
var(
	input devices.ElevInputDevice
)

func main() {

	// Given main function for going up and down and registering floors
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)
	fsm.FSM_init()

	// Channels for all the different inputs
	//drv_buttons := make(chan elevio.ButtonEvent)
	//drv_floors := make(chan int)
	//drv_obstr := make(chan bool)
	//drv_stop := make(chan bool)

	input := devices.Elevio_GetInputDevice()
    
	if(input.FloorSensor() == -1){
		fsm.FsmOnInitBetweenFloors();
	}

	for {
		// Request button
		var prevFloor [elevio.N_FLOORS][elevio.N_BUTTONS] int
		for f := 0; f < elevio.N_FLOORS; f++ {
			for b := 0; b < elevio.N_BUTTONS; b++ {
				v := input.RequestButton(elevio.ButtonType(b), f)
				if v != false && prevFloor[f][b] != v {
					fsm.FsmOnRequestButtonPress(f, elevio.ButtonType(b))
				}
				prevFloor[f][b] = v
			}
		}
		// Floor sensor
		var prevFloor int = -1
		f := input.FloorSensor()
		if f != -1 && f != prevFloor {
			fsm.FsmOnFloorArrival(f)
		}
		prevFloor = f

		// Timer
		if timer.TimerTimedOut() {
			timer.TimerStop()
			fsm.FsmOnDoorTimeout()
		}

		time.Sleep(time.Duration(25) * time.Millisecond)
	}
}