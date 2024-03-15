package main

import (
	backup "Sanntidsprogrammering/Elevator/backup"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	elevrun "Sanntidsprogrammering/Elevator/elevrun"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	hallassigner "Sanntidsprogrammering/Elevator/hallassigner"
)

type HardwareChannels struct {
	ChanButtons chan elevio.ButtonEvent
	ChanFloors  chan int
	ChanObstr   chan bool
}

type AssignerChannels struct {
	ChanElevatorTX           chan elevio.Elevator
	ChanElevatorRX           chan elevio.Elevator
	ChanMasterHallRequestsTX chan [elevio.N_FLOORS][2]bool
	ChanMasterHallRequestsRX chan [elevio.N_FLOORS][2]bool
}

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
	HardwareChannels := HardwareChannels{
		ChanButtons: make(chan elevio.ButtonEvent),
		ChanFloors:  make(chan int),
		ChanObstr:   make(chan bool),
	}

	AssignerChannels := AssignerChannels{
		ChanElevatorTX:           make(chan elevio.Elevator),
		ChanElevatorRX:           make(chan elevio.Elevator),
		ChanMasterHallRequestsTX: make(chan [elevio.N_FLOORS][2]bool),
		ChanMasterHallRequestsRX: make(chan [elevio.N_FLOORS][2]bool),
	}

	//Polling
	go elevio.PollButtons(HardwareChannels.ChanButtons)
	go elevio.PollFloorSensor(HardwareChannels.ChanFloors)
	go elevio.PollObstructionSwitch(HardwareChannels.ChanObstr)

	// Timer
	go fsm.CheckForTimeout()

	//Primary and backup
	backup.ListenForPrimary(HardwareChannels.ChanButtons, HardwareChannels.ChanFloors, HardwareChannels.ChanObstr,
		AssignerChannels.ChanElevatorTX, AssignerChannels.ChanMasterHallRequestsRX)
	go backup.SetToPrimary(AssignerChannels.ChanElevatorRX, AssignerChannels.ChanMasterHallRequestsTX)
	go hallassigner.RecieveAssignedOrders()

	elevrun.RunElevator(HardwareChannels.ChanButtons, HardwareChannels.ChanFloors, HardwareChannels.ChanObstr)

}