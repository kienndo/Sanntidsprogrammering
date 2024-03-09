package main
import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
	"encoding/json"
)

const (
	UDPPort     = 8000 // UDP port
	CheckPeriod = 1 * time.Second
)

// Define the struct
type Data struct {
	PrimaryAlive bool
	LastNumber   int
}

var (
	counter int
	OurData = Data{
		PrimaryAlive: true, 
		LastNumber: counter,
	}
)

func RunPrimary() {
	fmt.Println("Running as Primary")
	counter := 1

	// Får tak i siste counter-verdi
	if data, err := os.ReadFile("status.txt"); err == nil {
		if val, err := strconv.Atoi(string(data)); err == nil {
			counter = val
		}
	}
	StartBackupProcess()

	go UDPreceive()
	
	for {
		fmt.Println(counter)
		counter++

		if counter == 5 {
			counter = 1
		}

		// Sende counter-verdi til backup
		SendUDPMessage("localhost", UDPPort, strconv.Itoa(counter))

		time.Sleep(1 * time.Second)
	}
}

func UDPreceive(){
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

func RunBackup() {
	fmt.Println("Running as Backup")
	for {
		if PrimaryIsActive() {
			fmt.Println("Primary is active")
		} else {
			fmt.Println("Primary is inactive, taking over.") //Kjører en ny primary dersom den merker at det ikke er noen aktiv primary
			RunPrimary()
			return
		}
		time.Sleep(CheckPeriod)
	}
}

// Sjekker om primary er aktiv ved å sjekke om filen er oppdatert nylig
func PrimaryIsActive() bool {
	info, _ := os.Stat("status.txt")

	return time.Since(info.ModTime()) < 2*CheckPeriod
}

func StartBackupProcess() {
	// Åpner en ny terminal og kjører backup-prosessen
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "main.go", "backup")

	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start backup process:", err)
	}
}

func SendUDPMessage(host string, port int, message string) {

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

func main() {
	// Check if the program is run as a backup process
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "backup" {
		RunBackup()
	} else {
		RunPrimary()
	}
}