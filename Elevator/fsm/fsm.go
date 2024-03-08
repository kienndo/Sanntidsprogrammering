package fsm

import (
	devices "Sanntidsprogrammering/Elevator/devices"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	requests "Sanntidsprogrammering/Elevator/requests"
	timer "Sanntidsprogrammering/Elevator/timer"
	"fmt"
)

var (
	RunningElevator     elevio.Elevator
	outputDevice devices.ElevOutputDevice
	pair		 elevio.DirnBehaviourPair
)

// Initialize the Elevator with FSM
func FSM_init() {

	RunningElevator = elevio.Elevator{
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
			OutputDevice.RequestButtonLight(elevio.ButtonType(btn), floor, es.HallRequests[floor][btn] != false) //button
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

	switch RunningElevator.Behaviour {
	case elevio.EB_DoorOpen:
		if requests.ShouldClearImmediately(RunningElevator, btn_floor, btn_type) {
			timer.TimerStart(RunningElevator.Config.DoorOpenDuration)
		} else {
			RunningElevator.HallRequests[btn_floor][btn_type] = true
		}
	case elevio.EB_Moving:
		RunningElevator.HallRequests[btn_floor][btn_type] = true

	case elevio.EB_Idle:

		elevator.Request[btn_floor][btn_type] = 1
		pair = requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour
		switch pair.Behaviour {
		case elevio.EB_DoorOpen:
			OutputDevice.DoorLight(true)
			timer.TimerStart(RunningElevator.Config.DoorOpenDuration)
			RunningElevator = requests.ClearAtCurrentFloor(RunningElevator)
		case elevio.EB_Moving:
			OutputDevice.MotorDirection(elevio.MotorDirection(RunningElevator.Dirn))
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

	RunningElevator.Floor = newFloor

	OutputDevice.FloorIndicator(RunningElevator.Floor)

	switch RunningElevator.Behaviour {
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
	fmt.Println("\nNew state after arrival: %d\n", state)
}

func FsmOnDoorTimeout() {
	fmt.Printf("\n\nFsmOnDoorTimeout\n")
	fmt.Printf("slayer")

	switch RunningElevator.Behaviour {
	case elevio.EB_DoorOpen:
		pair := requests.ChooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = elevio.EB_Idle

		switch RunningElevator.Behaviour {
		case elevio.EB_DoorOpen:
			timer.TimerStart(elevator.Config.DoorOpenDuration)
			elevator = requests.ClearAtCurrentFloor(elevator)
			SetAllLights(elevator)
			elevator.Behaviour = elevio.EB_Idle
			break;

		case elevio.EB_Moving:
			elevator.Behaviour = elevio.EB_Idle
			//do nothing

		case elevio.EB_Idle:
			outputDevice.DoorLight(false)
			outputDevice.MotorDirection(elevio.MotorDirection(elevator.Dirn))
			elevator.Behaviour = elevio.EB_Idle
			break;
		}
		break;
	default:
		elevator.Behaviour = elevio.EB_Idle
		break;
	}

	fmt.Println("\nNew state after Timeout: %d\n", elevio.EbToString(elevator.Behaviour))
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
