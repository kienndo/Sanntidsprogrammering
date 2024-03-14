package hallassigner

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	bcast "Sanntidsprogrammering/Elevator/network/bcast"
	peers "Sanntidsprogrammering/Elevator/network/peers"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"sync"

	
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
	MasterHallRequests [elevio.N_FLOORS][2]bool
	AllElevators = make(map[string]HRAElevState)
	LastValidFloor int

	// Master recieve channel
	ChanElevatorTX = make(chan elevio.Elevator)
	ChanElevatorRX = make(chan elevio.Elevator)
	
	// Mutex
	HallRequestMutex sync.Mutex
	CostMutex sync.Mutex
	ElevatorMutex sync.Mutex

	// Port addresses
	ElevatorTransmitPort int = 1659

	// Cost function - input and output
	HRAOutput map[string][][2]bool
	Input HRAInput
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

	fmt.Println("NEW INPUT:" , Input)

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
	HRAOutput = *output
	
	fmt.Println("NEW OUTPUT:" , HRAOutput)
}	

func ButtonIdentifier(chanButtonRequests chan elevio.ButtonEvent, chanHallRequests chan elevio.ButtonEvent, chanCabRequests chan elevio.ButtonEvent) {
	for{
		select {
			case btnEvent := <-chanButtonRequests:
				if btnEvent.Button == elevio.BT_Cab{
					chanCabRequests <- btnEvent
				} else{
					chanHallRequests <- btnEvent
				}
			}
		}
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

func SendAssignedOrders(){
	for IP, NewHallOrders := range HRAOutput{
		jsonData, err := json.Marshal(NewHallOrders)
		if err != nil {
			return 
		}

		udpAddr, err := net.ResolveUDPAddr("udp", IP+":8080") // CHOOSE A NEW PORT
		if err != nil {
			return
		}

		conn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			return 
		}
		defer conn.Close()

		_, err = conn.Write(jsonData)
		if err != nil {
			return 
		}
	}
}

func RecieveNewAssignedOrders(){
	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil{
		fmt.Println("Error resolving UDP address: ", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil{
		fmt.Println("Error listening for UDP packets: ", err)
		return
	}
	defer conn.Close()

	for{
		buffer := make([]byte, 1024)
		n, _, _ := conn.ReadFromUDP(buffer)

		var AssignedHallRequests [][2]bool
		if err := json.Unmarshal(buffer[:n], &AssignedHallRequests); err != nil {
			fmt.Println("Error decoding JSON", err)
			continue
		}

		for i := 0; i < elevio.N_FLOORS; i++{
			for j:=0; j<2; j++{
				fsm.RunningElevator.Request[i][j]=AssignedHallRequests[i][j]
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
			}
		}

	}()

	for{
		select{
		case a:= <-ChanElevatorRX:
		
			UpdateHallRequests(a)
			//fmt.Println("MASTERHALLREQUESTS: ", MasterHallRequests)

			State := HRAElevState{
				Behavior: elevio.EbToString(a.Behaviour),
				Floor: a.Floor,
				Direction: elevio.ElevioDirnToString(a.Dirn),
				CabRequests: a.CabRequests[:],
			}
			//fmt.Println("NY IPADRESSE", IPaddress)
			ElevatorMutex.Lock()
			AllElevators[IPaddress] = State 
			ElevatorMutex.Unlock()
	
		}
	}
}
