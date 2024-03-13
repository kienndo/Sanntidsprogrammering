package watchdog

import (
	"time"
)

//const seconds = soemthing
func watchdogTimer (seconds int, elevatorUnavailable chan bool) {
	//timer! etter vi ikke har fått et signal 
	//ny timer, hvis dden kobler seg på etter 2 sek eks var det bare en liten disconnect
	watchdogTimer := time.NewTimer(time.Duration(seconds) * time.Second)
	for {
		select {
			case <- watchdogTimer.C:
				elevatorUnavailable <- true
				//send melding til master at denne heisen er død
				watchdogTimer.Reset(time.Duration(seconds) * time.Second) //restart timer
		}
	}
	//nettverksfeil, ikke en disconnect, 
	//heisen er fysisk disconnecta
	//master må få alle requests tilbake
}


func reassign(
	//channels
)