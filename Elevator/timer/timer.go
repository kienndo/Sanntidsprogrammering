package timer

// Direct translation from C to Golang, retrieved from https://github.com/TTK4145/Project-resources/tree/master/elev_algo:

import (
	
	"fmt"
	"time"
)

var (
	TimerEndTime float64
	TimerActive  int
)
// Starts a timer with a given duration 
func TimerStart(duration float64) {
	TimerEndTime = float64(time.Now().Unix()) + duration
	TimerActive = 1
	fmt.Println("Timer started")

}
// Stops the timer. Used to ensure the door is open for a given DoorOpenDuration
func TimerStop() {
	TimerActive = 0
}
// Checks if the timer has timed out and returns 1 if it has
func TimerTimedOut() int {
	if TimerActive != 0 && float64(time.Now().Unix()) > TimerEndTime {
		fmt.Println("Timed out")
		return 1
	}
	return 0
}

