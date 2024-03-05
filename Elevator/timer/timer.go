package timer

import (
	"time"
)

var (
	timerEndTime float64
	timerActive  int
)

type Timeval struct {
	Sec  int64
	Usec int64
}

func getWallTime() float64 {
	var wallTime float64
	currentTime := Timeval{Sec: time.Now().Unix(), Usec: int64(time.Now().UnixNano() / 1000)}
	wallTime = float64(currentTime.Sec) + float64(currentTime.Usec)*0.000001
	return wallTime
}

func TimerStart(duration float64) {
	timerEndTime = getWallTime() + duration
	timerActive = 1
}

func TimerStop() {
	timerActive = 0
}

func TimerTimedOut() bool {
	return float64(timerActive) != 0 && getWallTime() > timerEndTime
}