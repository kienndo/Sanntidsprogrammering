package main

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
	"fmt"
	backup "Sanntidsprogrammering/Elevator/backup"
	//bcast "Sanntidsprogrammering/Elevator/network/bcast"
)

func main() {

	//Initialization
	numFloors := 4
	elevio.Init("localhost:15657", numFloors)
	costfunctions.InitMasterHallRequests()
	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}

	//Creating channels
	ChanButtons := make(chan elevio.ButtonEvent)
	ChanFloors := make(chan int)
	ChanObstr := make(chan bool)
	ChanHallRequests := make(chan elevio.ButtonEvent)
	ChanCabRequests := make(chan elevio.ButtonEvent)
	
	//Polling 
	go elevio.PollButtons(ChanButtons)
	go elevio.PollFloorSensor(ChanFloors)
	go elevio.PollObstructionSwitch(ChanObstr)
	go costfunctions.ButtonIdentifier(ChanButtons,ChanHallRequests, ChanCabRequests)

	// Timer
	go fsm.CheckForTimeout()

	//Primary and backup
	backup.ListenForPrimary(ChanButtons, ChanFloors, ChanObstr)
	go backup.SetToPrimary()

	fsm.InitializeLights()

	for { // Put into function later?
		
		select {
		case a := <-ChanButtons:
			fmt.Printf("Order: %+v\n", a)
		
			fmt.Println("MASTERHALLREQUESTS", costfunctions.MasterHallRequests)
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)

			
		case a := <-ChanFloors:
			costfunctions.SetLastValidFloor(a)
			fmt.Printf("Floor: %+v\n", a)
			fsm.FsmOnFloorArrival(a)
			
		case a := <-ChanObstr:
			fmt.Printf("Obstructing: %+v\n", a)
			fsm.ObstructionIndicator = a
		}
	}
}


