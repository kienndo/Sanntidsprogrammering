package costfunctions

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"time"
	"sync"
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
	State1 HRAElevState
	State2 HRAElevState
	mutex sync.Mutex

	CurrentState = HRAElevState {

		Behavior:      "moving", //elevio.EbToString(HRAElevator.Behaviour),
		Floor:         1, //LastValidFloor, 
		Direction:     elevio.ElevioDirnToString(HRAElevator.Dirn),
		CabRequests:   HRAElevator.CabRequests, 
	}
	
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

	var Input = HRAInput{
		HallRequests: 	MasterHallRequests[:],
		States: map[string]HRAElevState{
			"one": CurrentState,
		},
	}
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
		go SendState(CurrentState, ":29501")

	} else {

		// Channel 2
		fmt.Println("sending to channel 2")
		go SendState(CurrentState, ":29502")

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

func SendState(state HRAElevState, addr string) {

	// Send states til master
	conn, err := net.Dial("udp","10.100.23.255"+addr)
	if err != nil {
		fmt.Println("Failed to dial UDP %vn", err)
		return
	}
	fmt.Println("kæser moren din kien")
	defer conn.Close()

	for {

	jsonData, err := json.Marshal(state)
	if err != nil {
		fmt.Println("failed to serialize data")
		return
	}

	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Println("failed to send state")
		return
	}

	// _, err := conn.Write([]byte("UDP connection funker på tide å kæse moren til kien"))
	// if err != nil {
	// 	fmt.Println("UDP connection funker ikke :(")
	// }
	time.Sleep(1*time.Second)
 }
}


func RecievingState(address string,state *HRAElevState) {

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println("failed to listen to udp mip")
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("failed to listen from udp")
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil{
			fmt.Println("failed to read")
			continue
		}

		// recievedString := string(buffer[:n])
		// fmt.Println(recievedString)
		var newState HRAElevState
		err = json.Unmarshal(buffer[:n],&newState)
		if err != nil {
			fmt.Println("failed to deserialize")
			continue
		}
		
		

		mutex.Lock()
		*state = newState
		fmt.Println("UDPSEND", newState)
	
		mutex.Unlock()


	}


}

func UpdateStates() {

	RecievingState(":29501", &State1)
	RecievingState(":29502", &State2)
	
}