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
	//chanAliveRX := make(chan bool)
	//chanAliveTX := make(chan bool)
	
	go elevio.PollButtons(chanButtons)
	go elevio.PollFloorSensor(chanFloors)
	go elevio.PollObstructionSwitch(chanObstr)

	go bcast.RunBroadcast()
	
	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}
	
	//backup.StartPrimary()
	//go func(){
	//bcast.Receiver(16564, chanAliveRX)
	//chanAliveTX <-backup.PrimaryIsActive()
	//bcast.Transmitter(16564, chanAliveTX)

    // if backup.PrimaryIsActive() {
    //     backup.RunPrimary()
		
    // } else {
    //     backup.RunBackup()
    // }
	// }()

	// go rutine som sjekker for primary
	go backup.ListenForPrimary()
	go backup.SetToPrimary()

	

	fsm.InitLights()

	for {
		
		select {
		case a := <-chanButtons:
			fmt.Printf("Order: %+v\n", a)
			
			//costfunctions.WhichButton(a) //vil ikke fungere med cab nå???
			//costfunctions.CostFunction() //printer ut noe stygt nå
	
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


