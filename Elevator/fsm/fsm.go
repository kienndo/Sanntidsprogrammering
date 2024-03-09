package fsm

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	requests "Sanntidsprogrammering/Elevator/requests"
	timer "Sanntidsprogrammering/Elevator/timer"
)

var (
	elevator = elevio.Elevator{
		Floor: -1,
		Dirn:  elevio.MD_Stop,
		Behaviour: elevio.EB_Idle,
		Request: [elevio.N_FLOORS][elevio.N_BUTTONS]int{{0, 0, 0}, 
														{0, 0, 0}, 
														{0, 0, 0}, 
														{0, 0, 0}},
		Config: elevio.Config{
			DoorOpenDuration:    3.0,
			ClearRequestVariant: elevio.CV_All,
		},
	}
	ObstructionIndicator bool
)
// func PrintState() {
// 	fmt.Println(StateToString(elevator.state))
// 	fmt.Println("Directelevion: ", elevator.dirn)
// }

func SetAllLights(es elevio.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if es.Request[floor][btn] != 0 {
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

func FsmOnRequestButtonPress(btn_Floor int, btn_type elevio.ButtonType) {
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
			elevator.Request[btn_Floor][btn_type] = 1
		}
	case elevio.EB_Moving:
		elevator.Request[btn_Floor][btn_type] = 1
	case elevio.EB_Idle:
		elevator.Request[btn_Floor][btn_type] = 1
		var pair elevio.DirnBehaviourPair = requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour
		switch pair.Behaviour {
		case elevio.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.TimerStart(elevator.Config.DoorOpenDuration)
			elevator = requests.ClearAtCurrentFloor(elevator)
		case elevio.EB_Moving:
			elevio.SetMotorDirection(elevator.Dirn)
		case elevio.EB_Idle:
			break
		}
	}
	SetAllLights(elevator)
}

func FsmOnFloorArrival(newFloor int) {
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
	for {
		if timer.TimerTimedOut() != 0 {
			timer.TimerStop()
			FsmOnDoorTimeout()
		}
	}
}

func FsmStopSignal(a bool){
	var prevState elevio.ElevatorBehaviour = elevator.Behaviour
	var prevDirn elevio.MotorDirection = elevator.Dirn
	
	for a == true {
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevator.Behaviour = elevio.EB_Idle
	}
	elevator.Behaviour = prevState
	elevator.Dirn = prevDirn
}

func FsmObstruction(a bool){

	if a == true && elevator.Behaviour == elevio.EB_DoorOpen{
		timer.TimerStart(elevator.Config.DoorOpenDuration)
	}
}