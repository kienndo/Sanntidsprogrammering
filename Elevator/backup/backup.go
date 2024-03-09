package backup

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
	udp "Sanntidsprogrammering/Elevator/udp"
)

const (
	UDPPort     = 8000 // UDP port
	checkPeriod = 1 * time.Second
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

	// Setter i gang Backup-prosessen
	StartBackupProcess()

	// Setter opp UDP
	go func() {
		udpAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", UDPPort))

		conn, _ := net.ListenUDP("udp", udpAddr)

		defer conn.Close()

		buf := make([]byte, 1024)
		for {
			n, _, _ := conn.ReadFromUDP(buf)

			fmt.Println("Received message from backup:", string(buf[:n]))
		}
	}()

	for {
		fmt.Println(counter)
		counter++

		if counter == 5 {
			counter = 1
		}
		// Oppdatere fil med siste counter-verdi
		os.WriteFile("status.txt", []byte(strconv.Itoa(counter)), 0666)

		// Sende counter-verdi til backup
		udp.SendUDP("localhost", UDPPort, strconv.Itoa(counter))

		time.Sleep(1 * time.Second)
	}
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
		time.Sleep(checkPeriod)
	}
}

// Sjekker om primary er aktiv ved å sjekke om filen er oppdatert nylig
func PrimaryIsActive() bool {
	info, _ := os.Stat("status.txt")

	return time.Since(info.ModTime()) < 2*checkPeriod
}

func StartBackupProcess() {
	// Åpner en ny terminal og kjører backup-prosessen
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "main.go", "backup")

	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start backup process:", err)
	}
}


// func main() {
// 	// Check if the program is run as a backup process
// 	args := os.Args[1:]
// 	if len(args) > 0 && args[0] == "backup" {
// 		runBackup()
// 	} else {
// 		runPrimary()
// 	}
// }