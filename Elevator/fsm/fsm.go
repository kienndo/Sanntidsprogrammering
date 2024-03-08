package fsm

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	requests "Sanntidsprogrammering/Elevator/requests"
	timer "Sanntidsprogrammering/Elevator/timer"
	"fmt"
	"time"
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
)

func InitLights() {
	elevio.SetDoorOpenLamp(false)
	SetAllLights(elevator)
}

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

func FsmOnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Dirn = elevio.MD_Down
	elevator.Behaviour = elevio.EB_Moving
}

func FsmOnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType) {
	fmt.Printf("\n\nFsmOnRequestButtonPress(%d, %s)\n", btn_floor, elevio.ElevioButtonToString(btn_type))

	switch elevator.Behaviour {
	case elevio.EB_DoorOpen:
		if requests.ShouldClearImmediately(elevator, btn_floor, btn_type) != 0 {
			timer.TimerStart(elevator.Config.DoorOpenDuration)
		} else {
			elevator.Request[btn_floor][btn_type] = 1
		}
	case elevio.EB_Moving:
		elevator.Request[btn_floor][btn_type] = 1

	case elevio.EB_Idle:

		elevator.Request[btn_floor][btn_type] = 1
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
	fmt.Printf("\n\nFsmOnFloorArrival%d\n", newFloor)

	elevator.Floor = newFloor

	elevio.SetFloorIndicator(elevator.Floor)

	switch elevator.Behaviour {
	case elevio.EB_Moving:
		if requests.ShouldStop(elevator) != 0 {
			
			elevio.SetMotorDirection(elevio.MotorDirection(elevio.MD_Stop))
			elevio.SetDoorOpenLamp(true)
			elevator := requests.ClearAtCurrentFloor(elevator)
			timer.TimerStart(elevator.Config.DoorOpenDuration)
			SetAllLights(elevator) //CAB LIGHTS?
			elevator.Behaviour = elevio.EB_DoorOpen
			
		}
	default:
		break
	}
}

func FsmOnDoorTimeout() {
	fmt.Printf("\n\nFsmOnDoorTimeout\n")

	switch elevator.Behaviour {
	case elevio.EB_DoorOpen:
		var pair elevio.DirnBehaviourPair = requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour

		switch elevator.Behaviour {
		case elevio.EB_DoorOpen:
			timer.TimerStart(elevator.Config.DoorOpenDuration)
			elevator = requests.ClearAtCurrentFloor(elevator)
			SetAllLights(elevator) // CAB???
			elevator.Behaviour = elevio.EB_Idle
		case elevio.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.Dirn)
			elevator.Behaviour = elevio.EB_Idle

		case elevio.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.Dirn)
			elevator.Behaviour = elevio.EB_Idle
		}
	default:
		elevator.Behaviour = elevio.EB_Idle
		//break
	}
}


func FsmCheckForDoorTimeout() {
	for {

		 if timer.TimerTimedOut() != 0 {
			timer.TimerStop()
			FsmOnDoorTimeout()
		 }
		 time.Sleep(10*time.Millisecond)
	}
}
