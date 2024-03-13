package main

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	//costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
	"fmt"
	backup "Sanntidsprogrammering/Elevator/backup"
	bcast "Sanntidsprogrammering/Elevator/network/bcast"
	costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
)

func main() {
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)
	costfunctions.InitMasterHallRequests()

	go fsm.CheckForTimeout()

	chanButtons := make(chan elevio.ButtonEvent)
	chanFloors := make(chan int)
	chanObstr := make(chan bool)
	//elevatorUnavailable := make(chan bool)
	
	go elevio.PollButtons(chanButtons)
	go elevio.PollFloorSensor(chanFloors)
	go elevio.PollObstructionSwitch(chanObstr)

	go bcast.RunBroadcast()

	go costfunctions.CostFunction() 
	
	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}

	go backup.ListenForPrimary()
	go backup.SetToPrimary()

	fsm.InitializeLights()

	go costfunctions.UpdateStates()


	for { // Put into function later?
		
		select {
		case a := <-chanButtons:
			fmt.Printf("Order: %+v\n", a)
		
			costfunctions.ButtonIdentifyer(a, costfunctions.ChanHallRequests) 
			
			costfunctions.CostFunction() 
	
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
			
		case a := <-chanFloors:
			costfunctions.SetLastValidFloor(a)
			fmt.Printf("Floor: %+v\n", a)
			fsm.FsmOnFloorArrival(a)
			

		case a := <-chanObstr:
			fmt.Printf("Obstructing: %+v\n", a)
			fsm.ObstructionIndicator = a

		case UpdateHallRequests := <-costfunctions.ChanHallRequests:
			
			costfunctions.MasterHallRequests[UpdateHallRequests.Floor][UpdateHallRequests.Button] = true
	
			fmt.Println("MasterHallRequests", costfunctions.MasterHallRequests)
				
		}
	}
}


