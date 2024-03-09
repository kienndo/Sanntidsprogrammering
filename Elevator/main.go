package main

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
	"fmt"
	//backup "Sanntidsprogrammering/Elevator/backup"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	go fsm.CheckForTimeout()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	
	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}
	//fmt.Println(master.Input)

	fsm.InitLights()
	//backup.RunBackup()
	//backup.RunPrimary()
	
	for {
		
		select {
		case a := <-drv_buttons:
			// Button signal
			fmt.Printf("%+v\n", a)
			
			costfunctions.WhichButton(a)
			costfunctions.CostFunction()
	
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
			
		case a := <-drv_floors:
			// Floor signal
			costfunctions.GetLastValidFloor(a)
			fmt.Printf("%+v\n", a)
			fsm.FsmOnFloorArrival(a)

		case a := <-drv_obstr:
			//Obstruction
			// Does not seem to want to keep the door open
			fmt.Printf("%+v\n", a)
			fmt.Printf("Obstructing?")
			fsm.ObstructionIndicator = a
			go fsm.FsmObstruction(fsm.ObstructionIndicator)
			
		case a := <-drv_stop: 
			// Does not keep going after the button is not pushed on
			//Stop button signal
			fmt.Printf("%+v\n", a)
			go fsm.FsmStopSignal(a) 
				
		}
	}
}