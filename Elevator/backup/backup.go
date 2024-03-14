package backup

import (
	hallassigner "Sanntidsprogrammering/Elevator/hallassigner"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	bcast "Sanntidsprogrammering/Elevator/network/bcast"
	localip "Sanntidsprogrammering/Elevator/network/localip"
	"fmt"
	"net"
	"time"
    "os"
)

// Functions for process pairs and indicating primary and backup

func ListenForPrimary(ChanButtons chan elevio.ButtonEvent, ChanFloors chan int, ChanObstr chan bool) {
    conn, err := net.ListenPacket("udp", ":29500")
    if err != nil {
        fmt.Println("Error listening")
    }
    defer conn.Close()

    buffer := make([]byte, 1024)
    conn.SetReadDeadline(time.Now().Add(5*time.Second))

    _, _, err = conn.ReadFrom(buffer) 
    if err != nil {
        return
    }

    fmt.Println("Backup started")

    timer := time.NewTimer(10*time.Second)
    go bcast.RunBroadcast(hallassigner.ChanElevator1, hallassigner.ElevatorTransmitPort) //Bytte navn på disse adressene
   
    // Run backup elevator too
    for {
        select {
        case <-timer.C:
            fmt.Println("Timeout expired, becoming primary")
            return
           
        case a := <-ChanButtons:
            fmt.Printf("Order: %+v\n", a)  
            fmt.Println("MASTERHALLREQUESTS", hallassigner.MasterHallRequests)
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
            timer.Reset(10 * time.Second)
        }
    }
}


func SetToPrimary() {

    time.Sleep(5*time.Second)

    conn, err := net.Dial("udp", "10.100.23.255:29500")
    if err != nil {
        fmt.Println("Error dialing UDP")
    }

    defer conn.Close()
    
    for {
        _, err := conn.Write([]byte("Primary alive"))
        if err != nil {
            fmt.Println("error sending in primary")
        }

        fmt.Println("Doing primarystuff")
        go hallassigner.MasterReceive()
        MasterIPAddress, _ := localip.LocalIP()
        MasterID := fmt.Sprintf("peer-%s-%d", MasterIPAddress, os.Getpid())
        hallassigner.AllElevators[MasterID] = hallassigner.HRAElevState{
                Behavior:   elevio.EbToString(fsm.RunningElevator.Behaviour),
                Floor:      fsm.RunningElevator.Floor, 
                Direction:  elevio.ElevioDirnToString(fsm.RunningElevator.Dirn),   
                CabRequests: fsm.RunningElevator.CabRequests[:],
            
        }
        hallassigner.CostFunction()

        time.Sleep(1*time.Second)
    }
}