package master

import (
	"fmt"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	"os/exec"
	"encoding/json"
	//fsm "Sanntidsprogrammering/Elevator/fsm"
)

const hraExecutable = "/home/student/Sanntidsprogrammering/Elevator/hall_request_assigner"

type HRAInput struct {
	HallRequests 	[][2]int							`json:"hallRequests"`
	States 			map[string]elevio.Elevator			 `json:"states"`
}

var(
	ElevatorCabRequests [elevio.N_FLOORS]int
	MasterHallRequests [elevio.N_FLOORS][2]int

	// Our three elevators
	Elevator1 = elevio.Elevator{
		Floor: -1,
		Dirn:  elevio.MD_Stop,
		Behaviour: elevio.EB_Idle,
		Request: [elevio.N_FLOORS][elevio.N_BUTTONS]int{{0, 0, 0}, 
														{0, 0, 0}, 
														{0, 0, 0}, 
														{0, 0, 0}},
		Config: elevio.Config{
			DoorOpenDuration:    3.0,
			ClearRequestVariant: elevio.CV_All,
		},
	}
	Elevator2 elevio.Elevator
	Elevator3 elevio.Elevator

	Input = HRAInput{
		HallRequests: 	[][2]int {{0, 0}, {1, 1}, {0, 0}, {0, 1}},
		States: map[string]elevio.Elevator{
			"one": {
				Behaviour:      Elevator1.Behaviour,
				Floor:          Elevator1.Floor,
				Dirn:           Elevator1.Dirn,
				Request:        Elevator1.Request,
			},
		},
	}
)


func CostFunction(){
	jsonBytes, err := json.Marshal(Input)
    if err != nil {
        fmt.Println("json.Marshal error: ", err)
        return
    }
    
    ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput() //"../hall_request_assigner/"+
    if err != nil {
        fmt.Println("exec.Command error: ", err)
        fmt.Println(string(ret))
        return
    }
    
    output := new(map[string][][2]bool)
    err = json.Unmarshal(ret, &output)
    if err != nil {
        fmt.Println("json.Unmarshal error: ", err)
        return
    }
        
    fmt.Printf("output: \n")
    for k, v := range *output {
        fmt.Printf("%6v :  %+v\n", k, v)
    }

}	

func WhichButton(btnEvent elevio.ButtonEvent,
	hallEvent chan elevio.ButtonEvent,
	cabEvent chan elevio.ButtonEvent) {

		switch {
		case btnEvent.Button == elevio.BT_Cab:
			fmt.Println("CAB", btnEvent)
			ElevatorCabRequests[btnEvent.Floor] = 1;
		case btnEvent.Button == elevio.BT_HallDown:
			fmt.Println("Hall",btnEvent)
			MasterHallRequests[btnEvent.Floor][btnEvent.Button] = 1;
		case btnEvent.Button == elevio.BT_HallDown:
			fmt.Println("Hall",btnEvent)
			MasterHallRequests[btnEvent.Floor][btnEvent.Button] = 1;
		default:
			break
		}
	}
