package main

import (
	//"Sanntidsprogrammering/Elevator/devices"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	timer "Sanntidsprogrammering/Elevator/timer"
	//"fmt"
	"time"
)

var (
	//input devices.ElevInputDevice
)

func main() {

	// Given main function for going up and down and registering floors
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)
	

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	

	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}
	// Request button
	var prevFloor = make([][]bool, elevio.N_FLOORS)
	for i := 0; i < elevio.N_FLOORS; i++ {
		prevFloor[i] = make([]bool, elevio.N_BUTTONS)
	}

	var previous int = -1

	for {

		

		for f := 0; f < elevio.N_FLOORS; f++ {
			for b := 0; b < elevio.N_BUTTONS; b++ {
				v := elevio.GetButton(elevio.ButtonType(b), f)
				if v != false && prevFloor[f][b] != v {
					fsm.FsmOnRequestButtonPress(f, elevio.ButtonType(b))
				}
				prevFloor[f][b] = v
			}
		}

		{
			// Floor sensor

			g := elevio.GetFloor()
			if g != -1 && g != previous {
				fsm.FsmOnFloorArrival(g)
			}
			previous = g
			

			if timer.TimerTimedOut() {
				timer.TimerStop()
				fsm.FsmOnDoorTimeout()
			}
		}
		time.Sleep(time.Duration(250) * time.Millisecond)
	}
}
