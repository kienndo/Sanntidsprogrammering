package hallassigner

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	bcast "Sanntidsprogrammering/Elevator/network/bcast"
	localip "Sanntidsprogrammering/Elevator/network/localip"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	"fmt"
	"time"
	"os"
	"sync"
)

var(
	// Initialization of variables
	MasterHallRequests [elevio.N_FLOORS][2]bool
	AllElevators = make(map[string]HRAElevState)
	LastValidFloor int
	PortMasterID int = 16667
	Input HRAInput

	// Mutex
	MasterHallRequestMutex sync.Mutex
	CostMutex sync.Mutex
	ElevatorMutex sync.Mutex
	HRAMutex sync.Mutex

)

// Send and recieve functions for new assigned orders
func SendAssignedOrders(HRAOutput map[string][][2]bool){
	
	ChanAssignedOrdersTX := make(chan map[string][][2]bool)
	go bcast.Transmitter(PortMasterID, ChanAssignedOrdersTX)

	go func() {
		for {
			HRAMutex.Lock()
			ChanAssignedOrdersTX <- HRAOutput
			HRAMutex.Unlock()

			time.Sleep(1 * time.Second)
		}
	}()
}
	
func RecieveAssignedOrders(){
	HRAMutex.Lock()
	HRAMutex.Unlock() 
	ChanAssignedOrdersRX := make(chan map[string][][2]bool)

	go bcast.Receiver(PortMasterID, ChanAssignedOrdersRX)

	for{
		select{
		case AssignedOrders:= <-ChanAssignedOrdersRX:
			for IP, AssignedHallRequests := range AssignedOrders{
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



		