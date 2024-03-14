package hallassigner

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	bcast "Sanntidsprogrammering/Elevator/network/bcast"
	peers "Sanntidsprogrammering/Elevator/network/peers"
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
	PortMasterID int = 16666

	// Master channels
	ChanElevatorTX = make(chan elevio.Elevator)
	ChanElevatorRX = make(chan elevio.Elevator)
	ChanMasterHallRequestsTX = make(chan [elevio.N_FLOORS][2]bool)
	ChanMasterHallRequestsRX = make(chan [elevio.N_FLOORS][2]bool)
	
	// Mutex
	HallRequestMutex sync.Mutex
	CostMutex sync.Mutex
	ElevatorMutex sync.Mutex

	// Port addresses
	ElevatorTransmitPort int = 1659
	MasterHallRequestsPort int = 1658

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
	
	for _, NewHallOrders := range HRAOutput{
		fmt.Println("Assigned orders: ", NewHallOrders)
		jsonData, err := json.Marshal(NewHallOrders)
		if err != nil {
			return 
		}

		udpAddr, err := net.ResolveUDPAddr("udp", "10.100.23.25:8080") // Sends to the given IP - address
		if err != nil {
			return
		}
		//fmt.Println("IP: ", udpAddr)

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
	// var ChanMasterIDRX chan peers.PeerUpdate

	// peers.Receiver(PortMasterID, ChanMasterIDRX)

	// for{
	// 	select{
	// 	case p := <- ChanMasterIDRX:
		addr, err := net.ResolveUDPAddr("udp", "10.100.23.24:8080")
		fmt.Println("IP: ")
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
	//}
//}

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
			fmt.Println("MASTERHALLREQUESTS: ", MasterHallRequests)

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

func MasterSendHallLights(){ // Må puttes i master, teste ved å bare sette masterhallrequests til en konstant matrise og bare teste at ting lyser
	
	ChanMasterHallRequestsTX <- MasterHallRequests
	bcast.Transmitter(MasterHallRequestsPort, ChanMasterHallRequestsTX)
}

func UpdateHallLights(){ // Må puttes i primary og master

	bcast.Receiver(MasterHallRequestsPort, ChanMasterHallRequestsRX)
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

