package main

import (

	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
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

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}

	fsm.InitLights() 

	for {
		
		select {
		case a := <-drv_buttons:
			// Button signal
			fmt.Printf("%+v\n", a)
			//io.SetButtonLamp(a.Button, a.Floor, true) hva er dette
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)

		case a := <-drv_floors:
			// Floor signal
			fmt.Printf("%+v\n", a)
			fsm.FsmOnFloorArrival(a)

		case a := <-drv_obstr:
			//Obstruction IMPLEMENT
			fmt.Printf("%+v\n", a)

		case a := <-drv_stop: //IMPLEMENT
			//Stop button signal
			fmt.Printf("%+v\n", a)
			//Turn all buttons off
			// for f := 0; f < numFloors; f++ {
			// 	for b := io.ButtonType(0); b < 3; b++ {
			// 		module.SetButtonLamp(b, f, false)
			// 	}
			// }
		}
	}
}