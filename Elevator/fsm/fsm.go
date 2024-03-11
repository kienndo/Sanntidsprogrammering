package fsm

import (
	"fmt"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	requests "Sanntidsprogrammering/Elevator/requests"
	timer "Sanntidsprogrammering/Elevator/timer"
	costfunctions "Sanntidsprogrammering/Elevator/costfunctions"
)

var (
	elevator = elevio.Elevator{
		Floor: 		-1,
		Dirn:  		elevio.MD_Stop,
		Behaviour: 	elevio.EB_Idle,
		Request: 	[elevio.N_FLOORS][elevio.N_BUTTONS]bool{{false, false, false}, 
															{false, false, false}, 
															{false, false, false}, 
															{false, false, false}},
		Config: 	elevio.Config {
					DoorOpenDuration:    3.0,
					ClearRequestVariant: elevio.CV_All,
					},
	}
	ObstructionIndicator bool
	elev = &elevator
)

func FSM (ch_ElevatorState chan<- elevio.Elevator,
		ch_ButtonState chan elevio.ButtonEvent,
		ch_FloorState chan int,
		ch_Obstruction chan bool) {
	//elev := &elevator
	// Initializing lights of the elevator
	// InitLights()
	ch_ElevatorState <- *elev // Sending the initial state of the elevator to the channel
	//CheckForTimeout()
	for {
		select {
		case ButtonSignal := <-ch_ButtonState:
			fmt.Printf("Order: %+v\n",ButtonSignal)
			costfunctions.WhichButton(ButtonSignal)
			costfunctions.CostFunction()
			FsmOnRequestButtonPress(ButtonSignal.Floor, 
									ButtonSignal.Button, 
									ch_ElevatorState)
		case FloorSignal := <-ch_FloorState:
			costfunctions.GetLastValidFloor(FloorSignal)
			fmt.Printf("Floor: %+v\n", FloorSignal)
			FsmOnFloorArrival	(FloorSignal,
								ch_ElevatorState)
		case ObstructionSignal := <-ch_Obstruction:
			fmt.Printf("Obstructing: %+v\n", ObstructionSignal)
			ObstructionIndicator = ObstructionSignal
		}
	}
}



func SetAllLights(es elevio.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if es.Request[floor][btn] { //!= false rewrite
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
			}
		}
	}
}

func InitLights() {
	elevio.SetDoorOpenLamp(false)
	SetAllLights(elevator)
}

func FsmOnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Dirn = elevio.MD_Down
	elevator.Behaviour = elevio.EB_Moving
}



func FsmOnRequestButtonPress(btn_Floor int, 
							btn_type elevio.ButtonType, 
							ch_ElevatorState chan<- elevio.Elevator) {
	switch elevator.Behaviour {
	case elevio.EB_DoorOpen:
		if requests.ShouldClearImmediately(elevator, btn_Floor, btn_type) != 0 {
			timer.TimerStart(elevator.Config.DoorOpenDuration)
		} else {
			if btn_type == 2 {
				//Btn_type += elevatornumber
			} else {
				//Update master matrix
			}
			elevator.Request[btn_Floor][btn_type] = true
		}
	case elevio.EB_Moving:
		elevator.Request[btn_Floor][btn_type] = true
	case elevio.EB_Idle:
		elevator.Request[btn_Floor][btn_type] = true
		var pair elevio.DirnBehaviourPair = requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour
		switch pair.Behaviour {
			case elevio.EB_DoorOpen:
				elevio.SetDoorOpenLamp(true)
				timer.TimerStart(elevator.Config.DoorOpenDuration)
				elevator = requests.ClearAtCurrentFloor(elevator)
				ch_ElevatorState <- *elev
			case elevio.EB_Moving:
				elevio.SetMotorDirection(elevator.Dirn)
				ch_ElevatorState <- *elev
			case elevio.EB_Idle:
				break
		}
	}
	SetAllLights(elevator)
}

func FsmOnFloorArrival(newFloor int, ch_ElevatorState chan<- elevio.Elevator) {
	elevator.Floor = newFloor
	elevio.SetFloorIndicator(elevator.Floor)

	switch elevator.Behaviour {
	case elevio.EB_Moving:
		if requests.ShouldStop(elevator) != 0 {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elevator = requests.ClearAtCurrentFloor(elevator)
			timer.TimerStart(elevator.Config.DoorOpenDuration)
			SetAllLights(elevator)
			elevator.Behaviour = elevio.EB_DoorOpen
			ch_ElevatorState <- *elev
		}
	default:
		break
	}
}

func FsmOnDoorTimeout() {
	switch elevator.Behaviour {
	case elevio.EB_DoorOpen:
		var pair elevio.DirnBehaviourPair = requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour

		switch elevator.Behaviour {
		case elevio.EB_DoorOpen:
			timer.TimerStart(elevator.Config.DoorOpenDuration)
			elevator = requests.ClearAtCurrentFloor(elevator)
			SetAllLights(elevator)
		case elevio.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.Dirn)
		case elevio.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.Dirn)
		}
	default:
		break
	}
}

func CheckForTimeout() {
	fmt.Println("jeg er her")
	for {
		if timer.TimerTimedOut() != 0 {
			fmt.Println("ikke timed out!")
			for ObstructionIndicator {
				FsmObstruction(ObstructionIndicator)
				fmt.Println("obstruction!")
			}
			timer.TimerStop()
			FsmOnDoorTimeout()
			fmt.Println("Timed out?")
		}
	}
}

func FsmObstruction(a bool){
	if a && elevator.Behaviour == elevio.EB_DoorOpen{
		timer.TimerStart(elevator.Config.DoorOpenDuration)
	}
}

