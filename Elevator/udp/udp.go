package udp

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"
	
)

const StatePort = "20011"
const DeletePort = "20012"
const AlivePort = "20013"

var State *net.UDPAddr
var Delete *net.UDPAddr
var Alive *net.UDPAddr

func UDPInit() {
	State, _ = net.ResolveUDPAddr("udp", ":" + StatePort)
	Delete, _ = net.ResolveUDPAddr("udp", ":" + DeletePort)
	Alive, _ = net.ResolveUDPAddr("udp", ":" + AlivePort)
}

func AliveSend() {

	// Create socket
	wrSocket, _ := net.DialUDP("udp", nil, Alive) 
	defer wrSocket.Close()

	// Create encoder
	encoder := gob.NewEncoder(wrSocket)

	message := "alive"

	for {
		//Encode and send
		err := encoder.Encode(message)
		if err != nil {
			fmt.Println("Klarer ikke å sende melding",err)
		}
	}

	time.Sleep(time.Second)

}

func SendUDP(host string, port int, message string) {
	udpAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Error dialing UDP:", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending UDP message:", err)
		return
	}
}

