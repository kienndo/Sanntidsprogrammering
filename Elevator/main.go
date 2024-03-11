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

	chanButtons := make(chan elevio.ButtonEvent)
	chanFloors := make(chan int)
	chanObstr := make(chan bool)
	
	go elevio.PollButtons(chanButtons)
	go elevio.PollFloorSensor(chanFloors)
	go elevio.PollObstructionSwitch(chanObstr)

	go bcast.RunBroadcast()
	
	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}
	
	backup.StartPrimary()
	

	fsm.InitLights()

	for {
		
		select {
		case a := <-chanButtons:
			fmt.Printf("Order: %+v\n", a)
			
			costfunctions.WhichButton(a)
			costfunctions.CostFunction()
	
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
			
		case a := <-chanFloors:

			costfunctions.GetLastValidFloor(a)
			fmt.Printf("Floor: %+v\n", a)
			fsm.FsmOnFloorArrival(a)
			

		case a := <-chanObstr:
			fmt.Printf("Obstructing: %+v\n", a)
			fsm.ObstructionIndicator = a
				
		}
	}
}


