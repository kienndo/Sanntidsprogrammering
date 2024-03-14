package costfunctions

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	bcast "Sanntidsprogrammering/Elevator/network/bcast"
	peers "Sanntidsprogrammering/Elevator/network/peers"
	//localip "Sanntidsprogrammering/Elevator/network/localip"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"sync"
	"time"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	
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
	MasterHallRequests [elevio.N_FLOORS][2]bool
	AllElevators = make(map[string]HRAElevState)
	LastValidFloor int
	//State1 HRAElevState
	//State2 HRAElevState

	// Master recieve channels
	ChanRecieveIP chan peers.PeerUpdate
	ChanRecieveElevator chan elevio.Elevator
	
	// Mutex
	HallRequestMutex sync.Mutex
	CostMutex sync.Mutex
	ElevatorMutex sync.Mutex

	ChanElevator1 = make(chan elevio.Elevator)
	ChanElevator2 = make(chan elevio.Elevator)
	Address1 int = 1659
	Address2 int = 1658

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

	select {
		case btnEvent := <-chanButtonRequests:
			if btnEvent.Button == elevio.BT_Cab{
				chanCabRequests <- btnEvent
			} else{
				chanHallRequests <- btnEvent
			}
		}
	}

func ChooseConnection() {
	// Sjekker om channel 1 er ledig
	conn, err := net.ListenPacket("udp",":29503")
	if err != nil {
		fmt.Println("Error listening to channel")
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(3*time.Second))
	_, _, err = conn.ReadFrom(buffer)

	if err != nil {

		// Channel 1
		fmt.Println("Sending to channel 1")
		go ChannelTaken()
		go bcast.RunBroadcast(ChanElevator1, Address1) //Kjøres bare en gang


	} else {

		// Channel 2
		fmt.Println("sending to channel 2")
		go bcast.RunBroadcast(ChanElevator2, Address2)
	}
	time.Sleep(1*time.Millisecond)
}

func ChannelTaken() {
	for {
		conn, err := net.Dial("udp", "10.100.23.255:29503")
		if err != nil {
			fmt.Println("Error dialing udp")
		}
		defer conn.Close()
		_, err = conn.Write([]byte("1"))

		time.Sleep(1*time.Second)
	}
}

func RecievingState(address string,state *elevio.Elevator) {

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println("failed to listen to udp mip")
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Failed to listen from UDP")
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil{
			fmt.Println("Failed to read")
			continue
		}

		var newState elevio.Elevator
		err = json.Unmarshal(buffer[:n],&newState)
		if err != nil {
			fmt.Println("Failed to deserialize")
			continue
		}
		
		CostMutex.Lock()
		*state = newState
		CostMutex.Unlock()
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
	// Sends the New hall order to the given IP-address
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
	ChanIP:= make(chan peers.PeerUpdate)
	go bcast.Receiver(Address1, ChanElevator2)
	go peers.Receiver(16666, ChanIP)
	var IPaddress string
	go func() {
		for{
			select{
			case p:= <-ChanIP:
				IPaddress = p.New //HVORDAN TAR JEG UT DENNE IPADRESSEN OG SENDER DEN UT AV FUNKSJONEN OG TIL SELECTEN UNDER
				fmt.Println("NY IPADRESSE", IPaddress)
			}
		}

	}()

	for{
		select{
		case a:= <-ChanElevator2:
		
			UpdateHallRequests(a)
			fmt.Println("MASTERHALLREQUESTS: ", MasterHallRequests) //Sjekk rosa markert kommentar i notability, kien

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
