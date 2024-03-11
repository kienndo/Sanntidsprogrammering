package main

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
	"fmt"
	backup "Sanntidsprogrammering/Elevator/backup"
	bcast "Sanntidsprogrammering/Elevator/network/bcast"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	go fsm.CheckForTimeout()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	go bcast.RunBroadcast()
	
	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}
	
	backup.StartPrimary()

	fsm.InitLights()

	for {
		
		select {
		case a := <-drv_buttons:
			fmt.Printf("Order: %+v\n", a)
			
			costfunctions.WhichButton(a)
			costfunctions.CostFunction()
	
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
			
		case a := <-drv_floors:

			costfunctions.GetLastValidFloor(a)
			fmt.Printf("Floor: %+v\n", a)
			fsm.FsmOnFloorArrival(a)
			

		case a := <-drv_obstr:
			fmt.Printf("Obstructing: %+v\n", a)
			fsm.ObstructionIndicator = a
				
		}
	}
}


