package master

import (
	"fmt"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	"os/exec"
	"encoding/json"
)

const hraExecutable = "/home/student/Sanntidsprogrammering/Elevator/hall_request_assigner"

type HRAElevState struct {
    Behavior    string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}


type HRAInput struct {
	HallRequests 	[][2]bool					`json:"hallRequests"`
	States 			map[string]HRAElevState		 `json:"states"`
}

var(
	MasterHallRequests [elevio.N_FLOORS][2]bool

	// Our three elevators
	Elevator1 = elevio.Elevator{
		Floor: -1,
		Dirn:  elevio.MD_Stop,
		Behaviour: elevio.EB_Idle,
		CabRequests: []bool {true, true, false, false},
	}
	
	Elevator2 elevio.Elevator
	Elevator3 elevio.Elevator

	Input = HRAInput{
		HallRequests: 	[][2]bool {{false, false}, {true, true}, {false, false}, {false, true}}, //må lage array for bare hallrequest
		States: map[string]HRAElevState{
			"one": {
				Behavior:      elevio.EbToString(Elevator1.Behaviour),
				Floor:         AvoidNegativeFloor(Elevator1), 
				Direction:     elevio.ElevioDirnToString(Elevator1.Dirn),
				CabRequests:    Elevator1.CabRequests, 
			},
		},
	}
)

func AvoidNegativeFloor(e elevio.Elevator) int {
	
	if e.Floor == -1{
		return 1
	}
	return e.Floor
}

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

func WhichButton(btnEvent elevio.ButtonEvent) {

		switch {
		case btnEvent.Button == elevio.BT_Cab:
			fmt.Println("CAB", btnEvent)
			Elevator1.CabRequests[btnEvent.Floor] = true;
		case btnEvent.Button == elevio.BT_HallDown:
			fmt.Println("Hall",btnEvent)
			MasterHallRequests[btnEvent.Floor][btnEvent.Button] = true;
		case btnEvent.Button == elevio.BT_HallDown:
			fmt.Println("Hall",btnEvent)
			MasterHallRequests[btnEvent.Floor][btnEvent.Button] = true;
		default:
			break
		}
	}
