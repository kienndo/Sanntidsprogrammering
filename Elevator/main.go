package main

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	hallassigner "Sanntidsprogrammering/Elevator/hallassigner"
	backup "Sanntidsprogrammering/Elevator/backup"
	localIP "Sanntidsprogrammering/Elevator/network/localip"
	"fmt"
	"os"
)

func main() {
	MasterIPAddress, _ := localIP.LocalIP()
	MasterID := fmt.Sprintf("%s:%d", MasterIPAddress, os.Getpid())
	fmt.Println("DETTE ER ID", MasterID)

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
	log.Println("Vi elsker Peter")


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
			fsm.ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKYag1fXxIBU6+lBG20N83iTfiAk6yJ01UMrEYenvu45 pepepopo
}


