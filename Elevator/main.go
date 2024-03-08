package main

import (
	//"Sanntidsprogrammering/Elevator/devices"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	//timer "Sanntidsprogrammering/Elevator/timer"
	"fmt"
	"time"
)

var (
	//input devices.ElevInputDevice
)

func main() {

	// Given main function for going up and down and registering floors
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)
	
	go fsm.FsmCheckForDoorTimeout()

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
	// var prevFloor = make([][]bool, elevio.N_FLOORS)
	// for i := 0; i < elevio.N_FLOORS; i++ {
	// 	prevFloor[i] = make([]bool, elevio.N_BUTTONS)
	// }

	// var previous int = -1

	for {
		select {
		case a := <-drv_buttons:
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
			//ElevatorModules.AddCabRequest(a.Floor, a.Button)
		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			fsm.FsmOnFloorArrival(a)
		
		time.Sleep(250 * time.Millisecond)
		}
	}
}