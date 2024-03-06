package fsm

import (
	devices "Sanntidsprogrammering/Elevator/devices"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	requests "Sanntidsprogrammering/Elevator/requests"
	timer "Sanntidsprogrammering/Elevator/timer"
	"fmt"
)

var (
	elevator     elevio.Elevator
	outputDevice devices.ElevOutputDevice
)

// Initialize the elevator with FSM
func FSM_init() {

	elevator = elevio.Elevator{
		Behaviour: elevio.EB_Idle,
		Config: elevio.Config{
			DoorOpenDuration:    3.0,
			ClearRequestVariant: elevio.CV_InDirn,
		},
	}
	fmt.Printf("FSM initialized")
	// Initialize outputDevice here
	outputDevice = devices.Elevio_GetOutputDevice()
	outputDevice.DoorLight(false)
}

func SetAllLights(es elevio.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			outputDevice.RequestButtonLight(elevio.ButtonType(btn), floor, es.Request[floor][btn] != 0) //button
		}
	}
}

func FsmOnInitBetweenFloors() {
	outputDevice.MotorDirection(elevio.MD_Down)
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
		elevator.Behaviour = pair.Behaviour
		switch pair.Behaviour {
		case elevio.EB_DoorOpen:
			outputDevice.DoorLight(true)
			timer.TimerStart(elevator.Config.DoorOpenDuration)
			elevator = requests.ClearAtCurrentFloor(elevator)
		case elevio.EB_Moving:
			outputDevice.MotorDirection(elevio.MotorDirection(elevator.Dirn))
		case elevio.EB_Idle:
			//do nothing
		}
	}
	SetAllLights(elevator)

	var state string = elevio.EbToString(elevator.Behaviour)
	fmt.Printf("\nNew state after button press: %d\n", state)
}

func FsmOnFloorArrival(newFloor int) {
	fmt.Printf("\n\nFsmOnFloorArrival%d\n", newFloor)

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
		break;
	}
	var state string = elevio.EbToString(elevator.Behaviour)
	fmt.Println("\nNew state after arrival: (%d)\n", state)
}

func FsmOnDoorTimeout() {
	fmt.Printf("\n\nFsmOnDoorTimeout(%d, %s)\n")
	fmt.Printf("slayer")

	switch elevator.Behaviour {
	case elevio.EB_DoorOpen:
		pair := requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour

		switch elevator.Behaviour {
		case elevio.EB_DoorOpen:
			timer.TimerStart(elevator.Config.DoorOpenDuration)
			elevator = requests.ClearAtCurrentFloor(elevator)
			SetAllLights(elevator)
			break;

		case elevio.EB_Moving:
			//do nothing

		case elevio.EB_Idle:
			outputDevice.DoorLight(false)
			outputDevice.MotorDirection(elevio.MotorDirection(elevator.Dirn))
			break;
		}
		break;
	default:
		break;
	}

	fmt.Println("\nNew state after Timeout: (%d)\n", elevio.EbToString(elevator.Behaviour))
}

func FSM_run(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr chan bool, drv_stop chan bool, numFloors int) {

	//Motor direction for testing
	var d elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(d)


	// Checks the state of the different input-channels
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	// Infinite loop that checks the state of the different input-channels and does something every time it gets a signal
	select {
	case a := <-drv_buttons:
		fmt.Printf("%+v\n", a)
		elevio.SetButtonLamp(a.Button, a.Floor, true)

	case a := <-drv_floors:
		fmt.Printf("%+v\n", a)
		if a == numFloors-1 {
			d = elevio.MD_Down
		} else if a == 0 {
			d = elevio.MD_Up
		}
		elevio.SetMotorDirection(d)

	case a := <-drv_obstr:
		fmt.Printf("%+v\n", a)
		if a {
			elevio.SetMotorDirection(elevio.MD_Stop)
		} else {
			elevio.SetMotorDirection(d)
		}

	case a := <-drv_stop:
		fmt.Printf("%+v\n", a)
		for f := 0; f < numFloors; f++ {
			for b := elevio.ButtonType(0); b < 3; b++ {
				elevio.SetButtonLamp(b, f, false)
			}
		}
	}

}
