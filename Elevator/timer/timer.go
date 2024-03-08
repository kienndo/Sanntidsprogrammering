package timer

import (
	"time"
	"fmt"
)

var (
	TimerEndTime float64
	TimerActive  int
)

func TimerStart(duration float64) {
	TimerEndTime = float64(time.Now().Unix()) + duration
	TimerActive = 1
	fmt.Println("Timer started")
}

func TimerStop() {
	TimerActive = 0
}

func TimerTimedOut() int {
	if TimerActive != 0 && float64(time.Now().Unix()) > TimerEndTime {
		fmt.Println("Timed out")
		return 1
	}
	return 0
}

