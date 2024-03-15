package main

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
	"fmt"
	backup "Sanntidsprogrammering/Elevator/backup"
	//watchdog "Sanntidsprogrammering/Elevator/watchdog"
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
	ElevatorEventsChannels := struct {
		ChanButtons chan elevio.ButtonEvent
		ChanFloors  chan int
		ChanObstr   chan bool
	}{
		ChanButtons: make(chan elevio.ButtonEvent),
		ChanFloors:  make(chan int),
		ChanObstr:   make(chan bool),
	}

	//Polling
	go elevio.PollButtons(ElevatorEventsChannels.ChanButtons)
	go elevio.PollFloorSensor(ElevatorEventsChannels.ChanFloors)
	go elevio.PollObstructionSwitch(ElevatorEventsChannels.ChanObstr)
	 
	//go costfunctions.HandleElevatorEvents(ElevatorEventsChannels)

	// Timer
	go fsm.CheckForTimeout()

	//Primary and backup
	//backup.ListenForPrimary(ChanButtons, ChanFloors, ChanObstr)
	go backup.SetToPrimary()

	fsm.InitializeLights()

	//go watchdog.WatchdogFunc(5, ElevatorUnavailable) //satt inn et random tall

	for { // Put into function later?

		select {
		case a := <-ElevatorEventsChannels.ChanButtons:
			fmt.Printf("Order: %+v\n", a)
			fmt.Println("MASTERHALLREQUESTS", costfunctions.MasterHallRequests)
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)

		case a := <-ElevatorEventsChannels.ChanFloors:
			costfunctions.SetLastValidFloor(a)
			fmt.Printf("Floor: %+v\n", a)
			fsm.FsmOnFloorArrival(a)

		case a := <-ElevatorEventsChannels.ChanObstr:
			fmt.Printf("Obstructing: %+v\n", a)
			fsm.ObstructionIndicator = a
		}
	}
}




