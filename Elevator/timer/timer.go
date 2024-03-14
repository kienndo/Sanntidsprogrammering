package timer

// Direct translation from C to Golang, retrieved from https://github.com/TTK4145/Project-resources/tree/master/elev_algo:

import (
	"time"
)

var (
	TimerEndTime float64
	TimerActive  int
)

func TimerStart(duration float64) {

	TimerEndTime = float64(time.Now().Unix()) + duration
	TimerActive = 1
}

func TimerStop() {

	TimerActive = 0
}

func TimerTimedOut() int {
	
	if TimerActive != 0 && float64(time.Now().Unix()) > TimerEndTime {
		return 1
	}
	return 0
}

