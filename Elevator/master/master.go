package master

import (
	"fmt"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	//"os"
)

var(
	CabRequests [elevio.N_FLOORS]int
	HallRequests [elevio.N_FLOORS][2]int
)

func WhichButton(btnEvent elevio.ButtonEvent,
	hallEvent chan elevio.ButtonEvent,
	cabEvent chan elevio.ButtonEvent) {

		switch {
		case btnEvent.Button == elevio.BT_Cab:
			fmt.Println("CAB", btnEvent)
			CabRequests[btnEvent.Floor] = 1;
		case btnEvent.Button == elevio.BT_HallDown:
			fmt.Println("Hall",btnEvent)
			HallRequests[btnEvent.Floor][btnEvent.Button] = 1;
		case btnEvent.Button == elevio.BT_HallDown:
			fmt.Println("Hall",btnEvent)
			HallRequests[btnEvent.Floor][btnEvent.Button] = 1;
		default:
			break
		}
	}


type master struct {
	//elevators map[string][]
}

