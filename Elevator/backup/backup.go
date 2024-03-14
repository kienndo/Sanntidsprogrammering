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

// Functions for process pairs and indicating primary and backup

func ListenForPrimary(ChanButtons chan elevio.ButtonEvent, ChanFloors chan int, ChanObstr chan bool) {

    
    conn, err := net.ListenPacket("udp", ":29502")
    if err != nil {
        fmt.Println("Error listening")
    }
    defer conn.Close()

    buffer := make([]byte, 1024)
    conn.SetReadDeadline(time.Now().Add(5*time.Second))

    _, _, err = conn.ReadFrom(buffer) 
    if err != nil {

        fmt.Println("Becoming primary")
        return
    }

    fmt.Println("Backup started")

    timer := time.NewTimer(2*time.Second)
    go bcast.RunBroadcast(hallassigner.ChanElevatorTX, hallassigner.ElevatorTransmitPort)
    //go hallassigner.RecieveNewAssignedOrders()
   
    // Run backup elevator too
    for {
        select {
        case <-timer.C:

            SleepRandomDuration()
            //fmt.Println("Timeout expired, becoming primary")
            ListenForPrimary(ChanButtons, ChanFloors, ChanObstr)
           
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
            timer.Reset(2* time.Second)
        }
    }
}


func SetToPrimary() {

    time.Sleep(5*time.Second)

    conn, err := net.Dial("udp", "10.100.23.255:29502") //Burde kanskje egt ikke kjøre en statisk IP-adresse?
    if err != nil {
        fmt.Println("Error dialing UDP")
    }

    defer conn.Close()

    
    for {
        _, err := conn.Write([]byte("Primary alive"))
        if err != nil {
            fmt.Println("Error sending in primary")
        }

        fmt.Println("Doing primarystuff")
        go hallassigner.MasterReceive()

        hallassigner.UpdateHallRequests(fsm.RunningElevator)
        fmt.Println("MASTERHALLREQUESTS: ", hallassigner.MasterHallRequests)
        MasterIPAddress, _ := localip.LocalIP()
        MasterID := fmt.Sprintf("%s:%d", MasterIPAddress, os.Getpid())
        hallassigner.AllElevators[MasterID] = hallassigner.HRAElevState{
                Behavior:   elevio.EbToString(fsm.RunningElevator.Behaviour),
                Floor:      fsm.RunningElevator.Floor, 
                Direction:  elevio.ElevioDirnToString(fsm.RunningElevator.Dirn),   
                CabRequests: fsm.RunningElevator.CabRequests[:],
            
        }
        hallassigner.CostFunction()
        //MasterSendID()
        hallassigner.SendAssignedOrders()

        time.Sleep(1*time.Second)
    }
}

func SleepRandomDuration() {

    rand.Seed(time.Now().UnixNano())
    duration := time.Duration(rand.Intn(5))*time.Second

    time.Sleep(duration)

}

func MasterSendID(){
    var MasterID string
    var ChanMasterIDTX chan string
    if MasterID == "" { 
		localIP, err := localip.LocalIP() 
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		MasterID = fmt.Sprintf("%s:%d", localIP, os.Getpid())
	}
    ChanMasterIDTX <- MasterID
    bcast.Transmitter(PortMasterID, ChanMasterIDTX)
}