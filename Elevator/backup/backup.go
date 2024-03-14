package backup

import (
	"Sanntidsprogrammering/Elevator/costfunctions"
	"Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	"Sanntidsprogrammering/Elevator/network/bcast"
	"fmt"
	"net"
	"time"
	//costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
)

// Function that uses a UDP connection to role out a primary/backup system.
// If a message is received within 5 seconds, the listener becomes a backup.
// Then, if no message is received within 10 seconds, the listener becomes primary.
func ListenForPrimary() {
    conn, err := net.ListenPacket("udp", ":29500")
    if err != nil {
        fmt.Println("Error listening")
    }
    defer conn.Close()

    buffer := make([]byte, 1024)
    conn.SetReadDeadline(time.Now().Add(5*time.Second))

    _, _, err = conn.ReadFrom(buffer) // If you read a message, you fall out of the loop
    if err != nil {
        return
    }

    fmt.Println("Backup started")
    // New timer
    timer := time.NewTimer(10*time.Second)

    // Begynner å sende states til primary
    //costfunctions.ChooseConnection()
    go bcast.RunBroadcast(costfunctions.ChanElevator1, costfunctions.Address1)
   
    

    for {
        select {
        case <-timer.C:
            fmt.Println("Timeout expired, becoming primary")
            return
           
        case a := <-ChanButtons:
            fmt.Printf("Order: %+v\n", a)  
            fmt.Println("MASTERHALLREQUESTS", costfunctions.MasterHallRequests)
            fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
           
        case a := <-ChanFloors:
            costfunctions.SetLastValidFloor(a)
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
    
    for { // Loops until it dies
        _, err := conn.Write([]byte("Primary alive"))
        if err != nil {
            fmt.Println("error sending in primary")
        }

        fmt.Println("Doing primarystuff")
        go costfunctions.MasterReceive()
        costfunctions.CostFunction()


        time.Sleep(1*time.Second)
    }
}