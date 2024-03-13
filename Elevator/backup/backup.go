package backup

import (

	"fmt"
	"net"
	"time"
    //costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
)

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

    timer := time.NewTimer(10*time.Second)

    // Begynner å sende states til primary
    

    for {
        select {
        case <-timer.C:
            fmt.Println("Timeout expired, becoming primary")
            return
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
        //go costfunctions.UpdateStates() // Burde denne egentlig ligge her

        time.Sleep(1*time.Second)
    }
}