package main

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	hallassigner "Sanntidsprogrammering/Elevator/hallassigner"
	backup "Sanntidsprogrammering/Elevator/backup"
	"fmt"

)

func main() {
	//Initialization
	numFloors := 4
	elevio.Init("localhost:15657", numFloors)
	hallassigner.InitMasterHallRequests()
	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}
	fsm.InitializeLights()

	//Creating channels
	ChanButtons := make(chan elevio.ButtonEvent)
	ChanFloors := make(chan int)
	ChanObstr := make(chan bool)
	
	//Polling 
	go elevio.PollButtons(ChanButtons)
	go elevio.PollFloorSensor(ChanFloors)
	go elevio.PollObstructionSwitch(ChanObstr)
	
	// Timer
	go fsm.CheckForTimeout()

	//Primary and backup
	backup.ListenForPrimary(ChanButtons, ChanFloors, ChanObstr)
	go backup.SetToPrimary()
	go hallassigner.RecieveNewAssignedOrders()

	// Run elevator
	for {
		
		select {
		case a := <-ChanButtons:
			fmt.Printf("Order: %+v\n", a)
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
			

		case a := <-ChanFloors:
			hallassigner.SetLastValidFloor(a)
			fmt.Printf("Floor: %+v\n", a)
			fsm.FsmOnFloorArrival(a)
			
		case a := <-ChanObstr:
			fmt.Printf("Obstructing: %+v\n", a)
			fsm.ObstructionIndicator = a
		}
}
}


