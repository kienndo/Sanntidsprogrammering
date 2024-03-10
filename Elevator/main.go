package main

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
	"fmt"
	backup "Sanntidsprogrammering/Elevator/backup"
	//localip "Sanntidsprogrammering/Elevator/network/localip"
)


func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	go fsm.CheckForTimeout()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	
	if elevio.GetFloor() == -1 {
		fsm.FsmOnInitBetweenFloors()
	}
	
	backup.StartPrimary()

	fsm.InitLights()

	for {
		
		select {
		case a := <-drv_buttons:
			// Button signal
			fmt.Printf("Order: %+v\n", a)
			
			costfunctions.WhichButton(a)
			costfunctions.CostFunction()
	
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
			
		case a := <-drv_floors:
			// Floor signal
			costfunctions.GetLastValidFloor(a)
			fmt.Printf("Floor: %+v\n", a)
			fsm.FsmOnFloorArrival(a)

		case a := <-drv_obstr:
			//Obstruction
			fmt.Printf("Obstructing: %+v\n", a)
			fsm.ObstructionIndicator = a
			
				
		}
	}
}

// For Network
// func main() {
// 	var id string
// 	flag.StringVar(&id, "id", "", "id of this peer") //flag.StringVar lagrer verdien av id i id variabelen 
// 	flag.Parse() 

// 	if id == "" { //hvis id er tom
// 		localIP, err := localip.LocalIP() //henter IP-adressen til enheten
// 		if err != nil {
// 			fmt.Println(err)
// 			localIP = "DISCONNECTED"
// 		}
// 		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid()) //id = peer-(IP-adresse)-(prosess-ID), os.Getpid() = henter prosess-ID, 
// 	}

// 	// Denne kanalen mottar oppdateringer om hvilke peers som er tilkoblet nettverket, inneholder Peers []string, New string og Lost []string
// 	peerUpdateCh := make(chan peers.PeerUpdate)

// 	// Denne kanalen aktiverer eller deaktiverer transmitteren
// 	peerTxEnable := make(chan bool) 
// 	go peers.Transmitter(15647, id, peerTxEnable) //Sender id til alle enheter på nettverket via UDP (fra peers.go)
// 	go peers.Receiver(15647, peerUpdateCh) //Mottar data fra alle enheter på nettverket via UDP og sender oppdateringer til peerUpdateCh (fra peers.go)

// 	helloTx := make(chan elevio.Elevator) //kanal for å sende HelloMsg -> sjekker argumenter og lagrer i selectCases og typeNames
// 	helloRx := make(chan elevio.Elevator) //kanal for å motta HelloMsg -> sjekker argumenter og lagrer i chansMap 

// 	go bcast.Transmitter(16569, helloTx) //Sender HelloMsg til alle enheter på nettverket via UDP (fra bcast.go)
// 	go bcast.Receiver(16569, helloRx) // Mottar HelloMsg fra alle enheter på nettverket via UDP (fra bcast.go)

// 	go func() {
// 		helloMsg := elevio.Elevator{
// 			Floor: -1,
// 			Dirn:  elevio.MD_Stop,
// 			Behaviour: elevio.EB_Idle,
// 			CabRequests: []bool {true, true, false, false},
// 		} 
// 		for {
// 			helloMsg.Floor++ // Inkrementerer Iter med 1 hvert sekund
// 			helloTx <- helloMsg //Sender helloMsg til helloTx-kanalen hvert sekund
// 			time.Sleep(1 * time.Second) //Venter 1 sekund
// 		}
// 	}()

// 	fmt.Println("Started")
// 	for {
// 		select {
// 		case p := <-peerUpdateCh: //hvis det kommer en melding på peerUpdateCh-kanalen
// 			fmt.Printf("Peer update:\n") //skriver ut "Peer update:"
// 			fmt.Printf("  Peers:    %q\n", p.Peers) //skriver ut "Peers: " og innholdet i p.Peers
// 			fmt.Printf("  New:      %q\n", p.New) //skriver ut "New: " og innholdet i p.New
// 			fmt.Printf("  Lost:     %q\n", p.Lost) //skriver ut "Lost: " og innholdet i p.Lost

// 		case a := <-helloRx: //hvis det kommer en melding på helloRx-kanalen
// 			fmt.Printf("Received: %#v\n", a) //skriver ut "Received: " og innholdet i a
// 		}
// 	}
// }