package udp

import (
	"fmt"
	"net"
)

func UDPrecieve() {

	address, _ := net.ResolveUDPAddr("udp", ":20011")

	// Listen to socket from address
	recSocket, _ := net.ListenUDP("udp", address)
	defer recSocket.Close()

	// Create decoder
	decoder := gob.NewDecoder(recSocket)

	// Create struct to store received data
	var receivedData Data

	for {
		// Decode data into struct
		err := decoder.Decode(&receivedData)
		if err != nil {
			fmt.Println("Error decoding struct:", err)
			return
		}

		//fmt.Println("Struct received:", receivedData)

		time.Sleep(time.Second)
	}
}

// Define the struct
type Data struct {
	PrimaryAlive bool
	LastNumber   int
}

func UDPsend(CountingNumber int) {
	// Create socket from address
	address, _ := net.ResolveUDPAddr("udp", ":20011") //MIN TERMINAL

	wrSocket, _ := net.DialUDP("udp", nil, address)
	defer wrSocket.Close()

	// Create encoder
	encoder := gob.NewEncoder(wrSocket)

	// Create struct instance to send
	dataToSend := Data{
		PrimaryAlive: true,
		LastNumber:   CountingNumber, // Example number
	}

	for {
		// Encode struct and send it
		err := encoder.Encode(dataToSend)
		if err != nil {
			fmt.Println("Error encoding struct:", err)
			return
		}

		//fmt.Println("Struct sent:", dataToSend)

		time.Sleep(time.Second)
	}
}