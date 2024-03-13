package watchdog

import (
	"time"
)


//const seconds = soemthing
func WatchdogFunc(seconds int, ElevatorUnavailable chan bool) {
	//timer! etter vi ikke har fått et signal 
	//ny timer, hvis dden kobler seg på etter 2 sek eks var det bare en liten disconnect
	watchdogTimer := time.NewTimer(time.Duration(seconds) * time.Second)
	for {
		select {
			case <- watchdogTimer.C:
				ElevatorUnavailable <- true
				//send melding til master at denne heisen er død
				watchdogTimer.Reset(time.Duration(seconds) * time.Second) //restart timer
			default: //
		}
	}
	//nettverksfeil, ikke en disconnect, 
	//heisen er fysisk disconnecta
	//master må få alle requests tilbake
}

// hvis elevatorUnavailable <- true, må det være en case et annet sted som tar seg av dette
// det som må skje da, er at heisen det gjelder må markeres som utilgjengelig og alle requests må sendes tilbake til master


