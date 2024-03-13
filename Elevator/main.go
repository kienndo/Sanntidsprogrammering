package main

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	//costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
	"fmt"
	backup "Sanntidsprogrammering/Elevator/backup"
	//bcast "Sanntidsprogrammering/Elevator/network/bcast"
)

func main() {

	//Initialization
	numFloors := 4
	elevio.Init("localhost:15657", numFloors)
	costfunctions.InitMasterHallRequests()
	//AllElevators := make(map[string]elevio.Elevator)
	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}

	//Creating channels
	ChanButtons := make(chan elevio.ButtonEvent)
	ChanFloors := make(chan int)
	ChanObstr := make(chan bool)
	ChanHallRequests := make(chan elevio.ButtonEvent)
	ChanCabRequests := make(chan elevio.ButtonEvent)
	//ChanUpdate := make(chan elevio.Elevator)

	//ChanElevator1 := make(chan elevio.Elevator)
	//ChanElevator2 := make(chan elevio.Elevator)

	//Polling 
	go elevio.PollButtons(ChanButtons)
	go elevio.PollFloorSensor(ChanFloors)
	go elevio.PollObstructionSwitch(ChanObstr)
	go costfunctions.ButtonIdentifier(ChanButtons,ChanHallRequests, ChanCabRequests)
	go costfunctions.UpdateHallRequests(ChanHallRequests)

	//go bcast.RunBroadcast(ElevatorMessageTX, ElevatorMessageRX, )
	go fsm.CheckForTimeout()
	go costfunctions.CostFunction() 
	go costfunctions.MasterRecieve()
	go costfunctions.ChooseConnection()

	//Primary and backup
	go backup.ListenForPrimary()
	go backup.SetToPrimary()

	fsm.InitializeLights()

	for { // Put into function later?
		
		select {
		case a := <-ChanButtons:
			fmt.Printf("Order: %+v\n", a)
		
			//costfunctions.CostFunction() 
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


