package main

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	master "Sanntidsprogrammering/Elevator/master"

	//requests "Sanntidsprogrammering/Elevator/requests"
	"fmt"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	go fsm.CheckForTimeout()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_hall := make(chan elevio.ButtonEvent)
	drv_cab := make(chan elevio.ButtonEvent)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}

	fsm.InitLights()
	//go fsm.FsmStopSignal()

	for {

		select {
		case a := <-drv_buttons:
			// Button signal
			fmt.Printf("%+v\n", a)
			
			master.WhichButton(a, drv_hall, drv_cab)
			fmt.Println("KOM SEG UT")
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)

		case a := <-drv_floors:
			// Floor signal
			fmt.Printf("%+v\n", a)
			fsm.FsmOnFloorArrival(a)

		case a := <-drv_obstr:
			//Obstruction
			// Does not seem to want to keep the door open
			fmt.Printf("%+v\n", a)
			fmt.Printf("Obstructing?")
			fsm.ObstructionIndicator = a
			go fsm.FsmObstruction(fsm.ObstructionIndicator)
			
		case a := <-drv_stop: 
			// Does not keep going after the button is not pushed on
			//Stop button signal
			fmt.Printf("%+v\n", a)
			go fsm.FsmStopSignal(a) 
				
		}
	}

}