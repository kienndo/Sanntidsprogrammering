package fsm

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	requests "Sanntidsprogrammering/Elevator/requests"
	timer "Sanntidsprogrammering/Elevator/timer"
)

// Initialization
var (
	RunningElevator = elevio.Elevator{
		Floor: -1,
		Dirn:  elevio.MD_Stop,
		Behaviour: elevio.EB_Idle,
		Request: [elevio.N_FLOORS][elevio.N_BUTTONS]bool{{false, false, false}, 
														{false, false, false}, 
														{false, false, false}, 
														{false, false, false}},
		Config: elevio.Config{
			DoorOpenDuration:    3.0,
			ClearRequestVariant: elevio.CV_All,
		},
	}
	ObstructionIndicator bool
)

// Direct translation from C to Golang, retrieved from https://github.com/TTK4145/Project-resources/tree/master/elev_algo
func SetAllLights(es elevio.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if es.Request[floor][btn] {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
			}
		}
	}
}

func InitLights() {
	elevio.SetDoorOpenLamp(false)
	SetAllLights(RunningElevator)
}

func FsmOnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	RunningElevator.Dirn = elevio.MD_Down
	RunningElevator.Behaviour = elevio.EB_Moving
}

func FsmOnRequestButtonPress(btn_Floor int, btn_type elevio.ButtonType) {
	switch RunningElevator.Behaviour {
	case elevio.EB_DoorOpen:
		if requests.ShouldClearImmediately(RunningElevator, btn_Floor, btn_type) != 0 {
			timer.TimerStart(RunningElevator.Config.DoorOpenDuration)
		} else {
			if btn_type == 2 {
				//Btn_type += elevatornumber HÆÆÆ
			} else {
				//Update master matrix
			}
			RunningElevator.Request[btn_Floor][btn_type] = true
		}
	case elevio.EB_Moving:
		RunningElevator.Request[btn_Floor][btn_type] = true
	case elevio.EB_Idle:
		RunningElevator.Request[btn_Floor][btn_type] = true
		var pair elevio.DirnBehaviourPair = requests.ChooseDirection(RunningElevator)
		RunningElevator.Dirn = pair.Dirn
		RunningElevator.Behaviour = pair.Behaviour
		switch pair.Behaviour {
		case elevio.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.TimerStart(RunningElevator.Config.DoorOpenDuration)
			RunningElevator = requests.ClearAtCurrentFloor(RunningElevator)
		case elevio.EB_Moving:
			elevio.SetMotorDirection(RunningElevator.Dirn)
		case elevio.EB_Idle:
			break
		}
	}
	SetAllLights(RunningElevator)
}

func FsmOnFloorArrival(newFloor int) {
	RunningElevator.Floor = newFloor
	elevio.SetFloorIndicator(RunningElevator.Floor)

	switch RunningElevator.Behaviour {
	case elevio.EB_Moving:
		if requests.ShouldStop(RunningElevator) != 0 {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			RunningElevator = requests.ClearAtCurrentFloor(RunningElevator)
			timer.TimerStart(RunningElevator.Config.DoorOpenDuration)
			SetAllLights(RunningElevator)
			RunningElevator.Behaviour = elevio.EB_DoorOpen
		}
	default:
		break
	}
}

func FsmOnDoorTimeout() {
	switch RunningElevator.Behaviour {
	case elevio.EB_DoorOpen:
		var pair elevio.DirnBehaviourPair = requests.ChooseDirection(RunningElevator)
		RunningElevator.Dirn = pair.Dirn
		RunningElevator.Behaviour = pair.Behaviour

		switch RunningElevator.Behaviour {
		case elevio.EB_DoorOpen:
			timer.TimerStart(RunningElevator.Config.DoorOpenDuration)
			RunningElevator = requests.ClearAtCurrentFloor(RunningElevator)
			SetAllLights(RunningElevator)
		case elevio.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(RunningElevator.Dirn)
		case elevio.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(RunningElevator.Dirn)
		}
	default:
		break
	}
}

func CheckForTimeout() {
	for {
		if timer.TimerTimedOut() != 0 {
			for ObstructionIndicator{
				FsmObstruction(ObstructionIndicator)
			}
			timer.TimerStop()
			FsmOnDoorTimeout()
		}
	}
}

func FsmObstruction(a bool){

	if a && RunningElevator.Behaviour == elevio.EB_DoorOpen{
		timer.TimerStart(RunningElevator.Config.DoorOpenDuration)
	}
}

