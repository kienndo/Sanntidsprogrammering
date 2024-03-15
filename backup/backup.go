package backup

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	hallassigner "Sanntidsprogrammering/Elevator/hallassigner"
	bcast "Sanntidsprogrammering/Elevator/network/bcast"
	localip "Sanntidsprogrammering/Elevator/network/localip"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

var PortMasterID int = 16666
var PrimaryPort int = 29502

// Functions for process pairs and indicating primary and backup

func ListenForPrimary(ChanButtons chan elevio.ButtonEvent, ChanFloors chan int, ChanObstr chan bool, ChanElevatorTX chan elevio.Elevator, ChanMasterHallRequestsRX chan [elevio.N_FLOORS][2]bool) {

	conn, err := net.ListenPacket("udp", fmt.Sprint(":", PrimaryPort)) // FIKSET
	if err != nil {
		fmt.Println("Error listening")
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // FIKSER ETTERPÅ

	_, _, err = conn.ReadFrom(buffer)
	if err != nil {

		fmt.Println("The system is ready, push a button :)")
		fmt.Println("Becoming primary")

		return
	}

	fmt.Println("Backup started")

	timer := time.NewTimer(2 * time.Second)
	go bcast.RunBroadcast(ChanElevatorTX, hallassigner.ElevatorPort)
    
	//go hallassigner.UpdateHallLights(ChanMasterHallRequestsRX)

	RunBackup(ChanButtons, ChanFloors, ChanObstr, conn, buffer, timer, ChanElevatorTX, ChanMasterHallRequestsRX)

}

func SetToPrimary(ChanElevatorRX chan elevio.Elevator, ChanMasterHallRequestsTX chan [elevio.N_FLOORS][2]bool) {

	time.Sleep(5 * time.Second)

	conn, err := net.Dial("udp", "10.100.23.255:29502") // FIKS ETTERPÅ MED TO HEISER
	if err != nil {
		fmt.Println("Error dialing UDP")
	}

	defer conn.Close()

	for {
		_, err := conn.Write([]byte("Primary alive"))
		if err != nil {
			fmt.Println("Error sending in primary")
		}

		go hallassigner.MasterReceive(ChanElevatorRX)
        hallassigner.InitMasterHallRequests()
		hallassigner.UpdateHallRequests(fsm.RunningElevator)
        fmt.Println("MASTER HALLREQUESTS: ", hallassigner.MasterHallRequests)
		MasterIPAddress, _ := localip.LocalIP()
		MasterID := fmt.Sprintf("%s:%d", MasterIPAddress, os.Getpid())
		hallassigner.AllElevators[MasterID] = hallassigner.HRAElevState{
			Behavior:    elevio.EbToString(fsm.RunningElevator.Behaviour),
			Floor:       fsm.RunningElevator.Floor,
			Direction:   elevio.ElevioDirnToString(fsm.RunningElevator.Dirn),
			CabRequests: fsm.RunningElevator.CabRequests[:],
		}
		hallassigner.CostFunction()
        
		//go hallassigner.MasterSendHallLights(ChanMasterHallRequestsTX)

		time.Sleep(1 * time.Second)
	}
}

func SleepRandomDuration() {

	rand.Seed(time.Now().UnixNano())
	duration := time.Duration(rand.Intn(2)) * time.Second

	time.Sleep(duration)
}

func RunBackup(ChanButtons chan elevio.ButtonEvent, ChanFloors chan int, ChanObstr chan bool, conn net.PacketConn, buffer []byte, timer *time.Timer, ChanElevatorTX chan elevio.Elevator, ChanMasterHallRequestsRX chan [elevio.N_FLOORS][2]bool) {

	for {
		select {
		case <-timer.C:

			SleepRandomDuration()
			ListenForPrimary(ChanButtons, ChanFloors, ChanObstr, ChanElevatorTX, ChanMasterHallRequestsRX)

		case a := <-ChanButtons:

			fmt.Printf("Order: %+v\n", a)
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
            
		case a := <-ChanFloors:

			hallassigner.SetLastValidFloor(a)
			fmt.Printf("Floor: %+v\n", a)
			fsm.FsmOnFloorArrival(a)

		case a := <-ChanObstr:

			fmt.Printf("Obstructing: %+v\n", a)
			fsm.ObstructionIndicator = a

		default:
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			_, _, err := conn.ReadFrom(buffer)
			if err != nil {
				continue
			}
			fmt.Println("Message received, restart timer")
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(2 * time.Second)
		}
	}

}
