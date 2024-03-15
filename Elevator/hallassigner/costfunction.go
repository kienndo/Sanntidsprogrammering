package hallassigner

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	bcast "Sanntidsprogrammering/Elevator/network/bcast"
	peers "Sanntidsprogrammering/Elevator/network/peers"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

type HRAElevState struct {
    Behavior    string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests 	[elevio.N_FLOORS][2]bool			`json:"hallRequests"`
	States 			map[string]HRAElevState		 		`json:"states"`
}

func InitMasterHallRequests(){
	MasterHallRequestMutex.Lock()
	MasterHallRequestMutex.Unlock()

	for i := 0; i<elevio.N_FLOORS; i++{
		for j:= 0; j<2; j++{
			
			MasterHallRequests[i][j] = false
		}
	}
}

var(
		// Port addresses
		ElevatorPort int = 1659
		MasterHallRequestsPort int = 1658

)

func SetLastValidFloor(ValidFloor int) {
	LastValidFloor = ValidFloor
}

func CostFunction(){
	MasterHallRequestMutex.Lock()
	MasterHallRequestMutex.Unlock()

	Input = HRAInput{
		HallRequests: MasterHallRequests,
		States: AllElevators,
	}

	hraExecutable := ""
    switch runtime.GOOS {
        case "linux":   hraExecutable  = "hall_request_assigner"
        case "windows": hraExecutable  = "hall_request_assigner.exe"
        default:        panic("OS not supported")
    }

    jsonBytes, err := json.Marshal(Input)
    if err != nil {
        fmt.Println("json.Marshal error: ", err)
        return
    }
    
    ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
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
	HRAOutput := *output
	go SendAssignedOrders(HRAOutput)
	
}	

func UpdateHallRequests(e elevio.Elevator){ 
	MasterHallRequestMutex.Lock()
	MasterHallRequestMutex.Unlock()

		for i:= 0; i<elevio.N_FLOORS; i++{
			for j:= 0; j<2; j++{
			if(e.HallRequests[i][j]){
				
				MasterHallRequests[i][j] = true 
				
			}
		}
	}
}

func MasterReceive(ChanElevatorRX chan elevio.Elevator){
	
	ChanRecieveID:= make(chan peers.PeerUpdate)
	var IPaddress string

	go bcast.Receiver(ElevatorPort, ChanElevatorRX)
	go peers.Receiver(15646, ChanRecieveID)

	go func() {
		for{
			select{
			case ID:= <-ChanRecieveID:
				IPaddress = ID.New
			}
		}

	}()

	for{
		select{
		case ElevUpdate:= <-ChanElevatorRX:
			InitMasterHallRequests()
			UpdateHallRequests(ElevUpdate)
			ElevatorCabs := MakeCabRequestsArray(ElevUpdate)

			State := HRAElevState{
				Behavior: elevio.EbToString(ElevUpdate.Behaviour),
				Floor: ElevUpdate.Floor,
				Direction: elevio.ElevioDirnToString(ElevUpdate.Dirn),
				CabRequests: ElevatorCabs,
			}

			ElevatorMutex.Lock()
			AllElevators[IPaddress] = State 
			ElevatorMutex.Unlock()
	
		}
	}
}

func MakeCabRequestsArray(e elevio.Elevator) []bool{

	CabRequests := make([]bool, elevio.N_FLOORS)
 
	 for i := 0; i < elevio.N_FLOORS; i++ {
		 CabRequests[i] = e.Request[i][2]
	 }
	 return CabRequests
 }
 

func MasterSendHallLights(ChanMasterHallRequestsTX chan [elevio.N_FLOORS][2]bool){
	MasterHallRequestMutex.Lock()
	MasterHallRequestMutex.Unlock()

	go bcast.Transmitter(MasterHallRequestsPort, ChanMasterHallRequestsTX)
	for{
		ChanMasterHallRequestsTX <- MasterHallRequests
	}
	
}

func UpdateHallLights(ChanMasterHallRequestsRX chan [elevio.N_FLOORS][2]bool){ 

	go bcast.Receiver(MasterHallRequestsPort, ChanMasterHallRequestsRX)
	for {
		select {
		case HallRequest := <-ChanMasterHallRequestsRX:
			for floor := 0; floor < elevio.N_FLOORS; floor++ {
				for btn := 0; btn < 2; btn++ {
					if HallRequest[floor][btn] == true {
						elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
						}
					}
				}
			}
		}
	}


