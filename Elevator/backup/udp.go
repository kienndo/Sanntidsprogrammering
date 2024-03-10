package backup

import (
	"encoding/gob"
	"fmt"
	"net"
	"encoding/json"
)

// VIL EGENTLIG HA DENNE FILEN I EN EGEN MAPPE, HVORDAN
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

	//time.Sleep(time.Second)

}
func SendUDPMessage(host string, port int, Data Data) { 

	// Create a UDP connection
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()


	// Serialize the struct to JSON
	data, err := json.Marshal(OurData)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Send the serialized struct over UDP
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func UDPreceive(){ // FLYTT TIL UDP
	addr := net.UDPAddr{
		Port: UDPPort,
		IP:   net.ParseIP("UDPPort"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Buffer to store received data
	buffer := make([]byte, 1024)

	// Receive data
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Deserialize the JSON back into a struct
	var receivedStruct Data
	err = json.Unmarshal(buffer[:n], &receivedStruct)
	if err != nil {
		fmt.Println(err)
		return
	}
   
   fmt.Printf("Received: %+v\n", receivedStruct)


}

