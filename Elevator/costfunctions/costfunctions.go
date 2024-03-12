package costfunctions

import (
	"fmt"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	"os/exec"
	"encoding/json"
	fsm "Sanntidsprogrammering/Elevator/fsm"
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
	LastValidFloor int
	HRAElevator = fsm.RunningElevator
	

	Input = HRAInput{
		HallRequests: 	[][2]bool {{false, false}, {true, true}, {false, false}, {false, true}}, //må lage array for bare hallrequest
		States: map[string]HRAElevState{
			"one": {
				Behavior:      elevio.EbToString(HRAElevator.Behaviour),
				Floor:         LastValidFloor, 
				Direction:     elevio.ElevioDirnToString(HRAElevator.Dirn),
				CabRequests:   HRAElevator.CabRequests, 
			},
		},
	}
)

func SetLastValidFloor(ValidFloor int) {
	LastValidFloor = ValidFloor
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

func ButtonIdentifyer(btnEvent elevio.ButtonEvent) {

		switch {
		case btnEvent.Button == elevio.BT_Cab:
			fmt.Println("CAB", btnEvent)
			HRAElevator.CabRequests[btnEvent.Floor] = true;
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
