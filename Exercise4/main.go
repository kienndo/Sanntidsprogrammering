
package main

import (
	"fmt"
	"os"
	"os/exec" // Make sure this import is included
	"strconv"
	"time"
)

const (
	statusFile  = "process_status.txt"
	checkPeriod = 1 * time.Second
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "backup" {
		runBackup()
	} else {
		runPrimary()
	}
}

func runPrimary() {
	fmt.Println("Running as Primary")
	counter := 1

	// Attempt to recover the last state if exists
	if data, err := os.ReadFile(statusFile); err == nil {
		if val, err := strconv.Atoi(string(data)); err == nil {
			counter = val
		}
	}

	// Spawn backup process
	startBackupProcess()

	for {
		fmt.Println(counter)
		counter++
		// Update the file with the latest counter value
		os.WriteFile(statusFile, []byte(strconv.Itoa(counter)), 0666)
		time.Sleep(1 * time.Second)
	}
}

func runBackup() {
	fmt.Println("Running as Backup")
	for {
		if primaryIsActive() {
			fmt.Println("Primary is active")
		} else {
			fmt.Println("Primary is inactive, taking over.")
			runPrimary()
			return
		}
		time.Sleep(checkPeriod)
	}
}

func primaryIsActive() bool {
	info, err := os.Stat(statusFile)
	if err != nil {
		return false
	}

	// Check if the file was updated recently
	return time.Since(info.ModTime()) < 2*checkPeriod
}

func startBackupProcess() {
	// Corrected command to ensure a new terminal window opens
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "pair_process.go", "backup")

	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start backup process:", err)
	}
}