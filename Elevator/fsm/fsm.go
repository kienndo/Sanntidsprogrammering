package fsm

import (
	devices "Sanntidsprogrammering/Elevator/devices"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	requests "Sanntidsprogrammering/Elevator/requests"
	timer "Sanntidsprogrammering/Elevator/timer"
	"fmt"
)


var (
	elevator elevio.Elevator
	outputDevice devices.ElevOutputDevice
	
)

// Initialize the elevator with FSM - but is it necessary to have an init in elevio too?
func FSM_init() {
	
	elevator = elevio.Elevator{
		Config: elevio.Config{
			DoorOpenDuration:   3.0,
			ClearRequestVariant: elevio.CV_InDirn,
		},
	}
	fmt.Printf("FSM initialized")
	// Initialize outputDevice here - men outputdevice gjør meg forvirret så skal prøve uten
	outputDevice = devices.Elevio_GetOutputDevice() 
}

func SetAllLights(es elevio.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			//outputDevice.RequestButtonLight(elevio.ButtonType(btn), floor, es.Requests[floor][btn] != 0) //button
		}
	}
}

func FsmOnInitBetweenFloors() {
	outputDevice.MotorDirection(elevio.MotorDirection(elevio.MD_Down))
	elevator.Dirn = elevio.MD_Down
	elevator.Behaviour = elevio.EB_Moving
}

func FsmOnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType) {
	fmt.Printf("\n\nFsmOnRequestButtonPress(%d, %s)\n", btn_floor, elevio.ElevioButtonToString(btn_type))
	

	switch elevator.Behaviour {
	case elevio.EB_DoorOpen:
		if requests.ShouldClearImmediately(elevator, btn_floor, btn_type) {
			timer.TimerStart(elevator.Config.DoorOpenDuration)
		} else {
			elevator.Request[btn_floor][btn_type] = 1
		}
	case elevio.EB_Moving:
		elevator.Request[btn_floor][btn_type] = 1

	case elevio.EB_Idle:
		elevator.Request[btn_floor][btn_type] = 1
		pair := requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour //skal være en del av structen
		switch pair.Behaviour {
		case elevio.EB_DoorOpen:
			outputDevice.DoorLight(true)
			timer.TimerStart(elevator.Config.DoorOpenDuration) //hmmm
			elevator = requests.ClearAtCurrentFloor(elevator)
		case elevio.EB_Moving:
			outputDevice.MotorDirection(elevio.MotorDirection(elevator.Dirn))
		case elevio.EB_Idle:
			//do nothing
		}
	}
	SetAllLights(elevator)
	fmt.Printf("\nNew state: \n")
	
}

func FsmOnFloorArrival(newFloor int) {
	fmt.Printf("\n\nFsmOnFloorArrival(%d, %s)\n", newFloor)
	

	elevator.Floor = newFloor

	outputDevice.FloorIndicator(elevator.Floor)

	switch elevator.Behaviour {
	case elevio.EB_Moving:
		if requests.ShouldStop(elevator) {
			outputDevice.MotorDirection(elevio.MotorDirection(elevio.MD_Stop))
			outputDevice.DoorLight(true)
			elevator := requests.ClearAtCurrentFloor(elevator)
			timer.TimerStart(elevator.Config.DoorOpenDuration)
			SetAllLights(elevator)
			elevator.Behaviour = elevio.EB_DoorOpen
		}
	}
	fmt.Println("\nNew state:")
	
}

func FsmOnDoorTimeout() {
	fmt.Printf("\n\nFsmOnDoorTimeout(%d, %s)\n")
	

	switch elevator.Behaviour {
	case elevio.EB_DoorOpen:
		pair := requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour

		switch elevator.Behaviour {
		case elevio.EB_DoorOpen:
			timer.TimerStart(elevator.Config.DoorOpenDuration)
			elevator := requests.ClearAtCurrentFloor(elevator)
			SetAllLights(elevator)

		case elevio.EB_Moving:

		case elevio.EB_Idle:
			outputDevice.DoorLight(false)
			outputDevice.MotorDirection(elevio.MotorDirection(elevator.Dirn))
		}
	}

	fmt.Println("\nNew state:")
	

}