package elevrun

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	hallassigner "Sanntidsprogrammering/Elevator/hallassigner"
	"fmt"
)

func RunElevator(ChanButtons chan elevio.ButtonEvent, ChanFloors chan int, ChanObstr chan bool){
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
			fsm.FsmObstruction(a)
		}
	}
}

