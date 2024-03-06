package timer

import (
	"time"
)

var (
	TimerEndTime float64
	TimerActive  int
)

type Timeval struct {
	Sec  int64
	Usec int64
}

func getWallTime() float64 {
	var WallTime float64
	CurrentTime := Timeval{Sec: time.Now().Unix(), Usec: int64(time.Now().UnixNano() / 1000)}
	WallTime = float64(CurrentTime.Sec) + float64(CurrentTime.Usec)*0.000001
	return WallTime
}

func TimerStart(duration float64) {
	TimerEndTime = getWallTime() + duration
	TimerActive = 1
}

func TimerStop() {
	TimerActive = 0
}

func TimerTimedOut() bool {
	return float64(TimerActive) != 0 && getWallTime() > TimerEndTime
}