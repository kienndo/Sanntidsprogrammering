package kien

import (
	//devices "Sanntidsprogrammering/Elevator/devices"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	"Sanntidsprogrammering/Elevator/requests"
	"time"
)

// gitt i main:
//numFloors := 4

//elevio.Init("localhost:15657", numFloors)
//fsm.FSM_init()
// if input.FloorSensor() == -1 {
// 	fsm.FsmOnInitBetweenFloors()
// }

var e elevio.Elevator

func main(){
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	// Channels for all the different inputs
	ChanButtons := make(chan elevio.ButtonEvent)	// Which button is pressed, but do i need it when i have hall and cab
	ChanFloors := make(chan int) 					// fordi hvilken etasje må kunne leses
	ChanObstr := make(chan bool) 					// obstruksjon, men bare på eller av
	ChanStop := make(chan bool) 					// stop, men bare på eller av
	ChanOrders := make(chan bool) 					// sende en ordre
	ChanHall := make(chan [elevio.N_FLOORS][2]bool) // sende hele den matrisen
	ChanCab := make(chan [elevio.N_FLOORS]bool) 	// sende hele av denne matrisen

	go StateMachine(ChanButtons, ChanFloors, ChanObstr, ChanStop, ChanOrders, ChanHall, ChanCab, numFloors)
}

func StateMachine(ChanButtons chan elevio.ButtonEvent, ChanFloors chan int, ChanObstr chan bool, ChanStop chan bool, ChanOrders chan bool, ChanHall chan [elevio.N_FLOORS][2]bool, ChanCab chan [elevio.N_FLOORS]bool, numFloors int) {
	
	fsm.FSM_init() // Needs to be redefined without OutputDevice and Config-config
	
	var obstr bool
	// Timer for DoorOpenDuration
	var DoorOpenTimer = time.NewTimer(time.Hour)
	
	// Configurating our elevator-object called RunningElevator, can this be done in FSM_init?
	fsm.RunningElevator = elevio.Elevator{
			Floor:       -1,
			Dirn:        elevio.MD_Stop,
			Behaviour:   elevio.EB_Idle,
			CabRequests: [elevio.N_FLOORS]bool{false, false, false, false},
			HallRequests: [elevio.N_FLOORS][2]bool{
				{false, false},
				{false, false},
				{false, false},
				{false, false}}, // matrixes to make up the requests
				Config: elevio.Config{
					DoorOpenDuration:    3.0,
					ClearRequestVariant: elevio.CV_InDirn, // Remove this cus idk what it does
				},
		}
		
		// Initialize the state machine with going down unless it is getting an order
		//elevio.SetMotorDirection(elevio.MD_Down)
		//fsm.RunningElevator.Dirn = elevio.MD_Down
		//fsm.RunningElevator.Behaviour = elevio.EB_Moving
		// okei skal jeg være helt ærlig skjønner jeg ikke hvorfor denne initialiseringen trengs fordi kan man ikke gå rett på state machine
		// Channels for all the different inputs
		
		// This for-loop basically just listens to all the channels to see if something happens
		for{
			select {
			
			case <-ChanFloors:
				// Is gonna indicate the floor
				elevio.SetFloorIndicator(fsm.RunningElevator.Floor)
				if requests.ShouldStop(fsm.RunningElevator) && fsm.RunningElevator.Behaviour == elevio.EB_Moving {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevio.SetDoorOpenLamp(true)
					e.Behaviour = elevio.EB_DoorOpen
					if !obstr {
						DoorOpenTimer.Reset(time.Duration(e.Config.DoorOpenDuration) * time.Second)
					}
				}	
			case <-ChanObstr: // do i need this in chan floors too?
				if obstr {
					// Stop timer
				}
				switch e.Behaviour {
				case elevio.EB_Idle: // do i need these too
				case elevio.EB_Moving: // do i need these too
				case elevio.EB_DoorOpen:
					if !obstr {
						// start timer again
					}
				}
			case <-ChanStop:
				// fuck må implementere selv
			case <-ChanHall: // Taking in an order from the hall
			switch e.Behaviour {
			case elevio.EB_DoorOpen:
			case elevio.EB_Moving:
			case elevio.EB_Idle:
				var tempDirnBehav elevio.DirnBehaviourPair = requests.ChooseDirection(e)
				e.Dirn = tempDirnBehav.Dirn
				e.Behaviour = tempDirnBehav.Behaviour
				switch e.Behaviour {
				case elevio.EB_Idle:
				case elevio.EB_DoorOpen:
					elevio.SetDoorOpenLamp(true)
					//set timer

				case elevio.EB_Moving:
					elevio.SetMotorDirection(e.Dirn)
				}
			}
			case CabButtonEvent := <-ChanCab:
				//e.CabRequests[] = true LØSE DENNE
				elevio.SetButtonLamp(elevio.BT_Cab, CabButtonEvent.Floor, true)
			
				switch e.Behaviour {
				case elevio.EB_DoorOpen:
				case elevio.EB_Moving:
				case elevio.EB_Idle:
					var tempDirnBehav elevio.DirnBehaviourPair = requests.ChooseDirection(e)
					e.Dirn = tempDirnBehav.Dirn
					e.Behaviour = tempDirnBehav.Behaviour
					switch e.Behaviour {
					case elevio.EB_Idle:
					case elevio.EB_DoorOpen:
						elevio.SetDoorOpenLamp(true)
						DoorOpenTimer.Reset(time.Duration(e.Config.DoorOpenDuration) * time.Second)

					case elevio.EB_Moving:
						elevio.SetMotorDirection(e.Dirn)
					}
				}
			}
		}
}

// For å skille mellom hvilke knapper som er trykket på
func BtnEventSplitter(btnEvent chan elevio.ButtonEvent,
	hallEvent chan elevio.ButtonEvent,
	cabEvent chan elevio.ButtonEvent) {
	for {
		select {
		case event := <-btnEvent:
			if event.Button == elevio.BT_Cab {
				cabEvent <- event
			} else {
				hallEvent <- event
			}
		}
	}
}

