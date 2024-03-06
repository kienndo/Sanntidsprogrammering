package main

import (
	"Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	//"fmt"
)

func main() {

	// Given main function for going up and down and registering floors
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)
	fsm.FSM_init()

	// Channels for all the different inputs
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	// Using state machine to run the elevator
	for {
	fsm.FSM_run(drv_buttons, drv_floors, drv_obstr, drv_stop, numFloors)
	}
}
