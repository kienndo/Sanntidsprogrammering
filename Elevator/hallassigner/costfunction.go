package hallassigner

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	bcast "Sanntidsprogrammering/Elevator/network/bcast"
	peers "Sanntidsprogrammering/Elevator/network/peers"
	"time"
	"os"
	localip "Sanntidsprogrammering/Elevator/network/localip"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"reflect"
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

var(
	// Initialization of variables
	//MasterID string
	MasterHallRequests [elevio.N_FLOORS][2]bool
	AllElevators = make(map[string]HRAElevState)
	LastValidFloor int
	PortMasterID int = 16666
	Input HRAInput

	// Master channels
	ChanElevatorTX = make(chan elevio.Elevator)
	ChanElevatorRX = make(chan elevio.Elevator)
	ChanMasterHallRequestsTX = make(chan [elevio.N_FLOORS][2]bool)
	ChanMasterHallRequestsRX = make(chan [elevio.N_FLOORS][2]bool)
	
	// Mutex
	HallRequestMutex sync.Mutex
	CostMutex sync.Mutex
	ElevatorMutex sync.Mutex
	MasterMutex sync.Mutex
	HRAMutex sync.Mutex

	// Port addresses
	ElevatorTransmitPort int = 1659
	MasterHallRequestsPort int = 1658

	watchdogTimer = time.NewTimer(time.Duration(5) * time.Second)

)

func InitMasterHallRequests(){
	for i := 0; i<elevio.N_FLOORS; i++{
		for j:= 0; j<2; j++{
			MasterHallRequests[i][j] = false
		}
	}
}

func SetLastValidFloor(ValidFloor int) {
	LastValidFloor = ValidFloor
}

func CostFunction(){

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
	InitMasterHallRequests()
	
}	

func UpdateHallRequests(e elevio.Elevator){ 
		for i:= 0; i<elevio.N_FLOORS; i++{
			for j:= 0; j<2; j++{
			if(e.Request[i][j]){
				HallRequestMutex.Lock()
				MasterHallRequests[i][j] = true 
				HallRequestMutex.Unlock()	
			}
		}
	}
}

func SendAssignedOrders(HRAOutput map[string][][2]bool){
	ChanAssignedOrders := make(chan map[string][][2]bool)
	
	go bcast.Transmitter(16667, ChanAssignedOrders) // Denne porten har et navn
	go func() {
		for {
			HRAMutex.Lock()
			ChanAssignedOrders <- HRAOutput
			HRAMutex.Unlock()

			time.Sleep(1 * time.Second)
		}
	}()
}
	
func RecieveNewAssignedOrders(){
	HRAMutex.Lock()
	HRAMutex.Unlock()
	ChanAssignedOrdersRec := make(chan map[string][][2]bool)

	go bcast.Receiver(16667, ChanAssignedOrdersRec)

	for{
		select{
		case p:= <-ChanAssignedOrdersRec:
			for IP, AssignedHallRequests := range p{
				LocalIP, _ := localip.LocalIP()
				if IP == fmt.Sprintf("%s:%d", LocalIP, os.Getpid()){
					for i := 0; i < elevio.N_FLOORS; i++{
							for j:=0; j<2; j++{
								if AssignedHallRequests[i][j] == true{
								fsm.RunningElevator.Request[i][j]=true
								}
							}
						}

				}
			}
		}
	}
}

func MasterReceive(){
	
	ChanRecieveIP:= make(chan peers.PeerUpdate)
	var IPaddress string

	go bcast.Receiver(ElevatorTransmitPort, ChanElevatorRX)
	go peers.Receiver(15646, ChanRecieveIP)

	go func() {
		for{
			select{
			case p:= <-ChanRecieveIP:
				IPaddress = p.New
				if len(p.Lost) > 0 {
					watchdogTimer.Reset(time.Duration(5) * time.Second)
					go func() {
						for {
							if reflect.DeepEqual(p.Lost, p.New) {
								watchdogTimer.Stop()
								fmt.Println("p.Lost has become p.New before timer expired")
                				return
							}
							time.Sleep(time.Millisecond * 100)
						}
					}()
					select {
					case <-watchdogTimer.C:
						fmt.Println("Elevator is deaddddd")
						unavailableElevator := p.Lost[0]
						newAllElevators := make(map[string]HRAElevState) 
						ElevatorMutex.Lock()
						for ID, elevator := range AllElevators {
							if ID != unavailableElevator {
								newAllElevators[ID]= elevator 
							}
						}
						peerUpdate := peers.PeerUpdate{
							Peers:       p.Peers,
							New:         p.New,
							Lost:        []string{},
							Unavailable: []string{p.Lost[0]},
						}
						peerUpdateCh := peers.PeerUpdateCh
						peerUpdateCh <- peerUpdate 
						AllElevators = newAllElevators
						ElevatorMutex.Unlock()
					default:
						// do nothing	
					}		
				}
			}
		}
	}()

	for{
		select{
		case a:= <-ChanElevatorRX:
		
			UpdateHallRequests(a)

			State := HRAElevState{
				Behavior: elevio.EbToString(a.Behaviour),
				Floor: a.Floor,
				Direction: elevio.ElevioDirnToString(a.Dirn),
				CabRequests: a.CabRequests[:],
			}

			ElevatorMutex.Lock()
			AllElevators[IPaddress] = State 
			ElevatorMutex.Unlock()
		}
	}
}

func MasterSendHallLights(){
	go bcast.Transmitter(MasterHallRequestsPort, ChanMasterHallRequestsTX)
	for{
		ChanMasterHallRequestsTX <- MasterHallRequests
	}
	
}

func UpdateHallLights(){ 
	go bcast.Receiver(MasterHallRequestsPort, ChanMasterHallRequestsRX)
	for {
		select {
		case a := <-ChanMasterHallRequestsRX:
			for floor := 0; floor < elevio.N_FLOORS; floor++ {
				for btn := 0; btn < 2; btn++ {
					if a[floor][btn] == true {
						elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
						}
					}
				}
			}
		}
	}

// Vent hvor skal jeg cleare
func ClearAssignedOrders(){
	for i:=0; i<elevio.N_FLOORS; i++{
		for j:=0; j<2; j++{
			if(fsm.RunningElevator.Request[i][j]){
				fsm.RunningElevator.Request[i][j] = false
			}
		}
	}
}



