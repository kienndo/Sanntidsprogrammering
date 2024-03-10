package main

import (
	"fmt"
	"os/exec"
	"time"
	udp "Sanntidsprogrammering/Elevator/backup"
	"os"
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

	//MÅ FÅ TAK I
	StartBackupProcess()

	go udp.UDPreceive()
	
	for{	// Sende counter-verdi til backup
	udp.SendUDPMessage("localhost", UDPPort, udp.OurData)
	}

	time.Sleep(1 * time.Second)
}


// Sjekker om primary er aktiv ved å sjekke om filen er oppdatert nylig
func PrimaryIsActive() bool {
	return OurData.PrimaryAlive == true
}

func StartBackupProcess() {
	// Åpner en ny terminal og kjører backup-prosessen
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "main.go", "backup")

	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start backup process:", err)
	}
}

func RunBackup() {
	fmt.Println("Running as Backup")
	for {
		if PrimaryIsActive() {
			fmt.Println("Primary is active")
		} else {
			fmt.Println("Primary is inactive, taking over.") 
			RunPrimary()
			return
		}
		time.Sleep(CheckPeriod)
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