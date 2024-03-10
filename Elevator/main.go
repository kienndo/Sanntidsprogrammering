package main

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
	"fmt"
	backup "Sanntidsprogrammering/Elevator/backup"
	"os"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	go fsm.CheckForTimeout()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	
	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}
	
	backup.StartPrimary()

	fsm.InitLights()

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "backup" {
		backup.RunBackup()
	} else {
		backup.RunPrimary()
	}

	
	for {
		
		select {
		case a := <-drv_buttons:
			// Button signal
			fmt.Printf("Order: %+v\n", a)
			
			costfunctions.WhichButton(a)
			costfunctions.CostFunction()
	
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
			
		case a := <-drv_floors:
			// Floor signal
			costfunctions.GetLastValidFloor(a)
			fmt.Printf("Floor: %+v\n", a)
			fsm.FsmOnFloorArrival(a)

		case a := <-drv_obstr:
			//Obstruction
			fmt.Printf("Obstructing: %+v\n", a)
			fsm.ObstructionIndicator = a
			
				
		}
	}
}