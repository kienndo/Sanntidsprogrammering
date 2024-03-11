package backup

import(
	// "net"
	// "fmt"
	// "time"
	// "math/rand"
	// bcast "Sanntidsprogrammering/Elevator/network/bcast"
	//localip "Sanntidsprogrammering/Elevator/network/localip"
	//"os"
	//peers "Sanntidsprogrammering/Elevator/network/peers"
)

func StartPrimary() {
	
	
	
	//go PrimaryIsActive()

	
	// go bcast.Transmitter(156476, ChanAliveTX)

	// ChanAliveRX := make(chan bool)
	
    // go bcast.Receiver(156476, ChanAliveRX)
	
    //IfPrimaryAlive := <-ChanAliveRX
    // args := os.Args[1:]
    // if len(args) > 0 && args[0] == "backup" {
    //     RunBackup()
    // } else {
    //     RunPrimary()
    // }
}

	// addr, err := net.ResolveUDPAddr("udp", "localhost:15657")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// conn, err := net.ListenUDP("udp", addr)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer conn.Close()

	// // Initialize the random seed
    // rand.Seed(time.Now().UnixNano())

    // Generate a random duration between 0 and 10 seconds
    // randomDuration := time.Millisecond * time.Duration(rand.Intn(10))

	// buffer := make([]byte, 1024)

    // // Set the read deadline from the current time plus the random duration
    // // err = conn.SetReadDeadline(time.Now().Add(randomDuration))
	// // _,_, err = conn.ReadFromUDP(buffer)