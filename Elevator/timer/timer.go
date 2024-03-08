package timer

import (
	"time"
	"fmt"
	//fsm "Sanntidsprogrammering/Elevator/fsm"
)

var (
	TimerEndTime float64
	TimerActive  int
)

// type Timeval struct {
// 	Sec  int64
// 	Usec int64
// }

// func getWallTime() float64 {
// 	var WallTime float64
// 	CurrentTime := Timeval{Sec: time.Now().Unix(), Usec: int64(time.Now().UnixNano() / 1000)}
// 	WallTime = float64(CurrentTime.Sec) + float64(CurrentTime.Usec)*0.000001
// 	return WallTime
// }

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

