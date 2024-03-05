package devices

import (
	elevio "Sanntidsprogrammering/Elevator/elevio" 
)

/*Creating structs for if the elevator is an input device og an output device. If it is a input device, 
it is usually taking in a signal, like  for example which floor am i on or has the stop button been pushed.
The output device will trigger things such as motor direction, lights and indicators.
*/

type ElevInputDevice struct {
	FloorSensor   func() int
	RequestButton func(elevio.ButtonType, int) bool
	StopButton    func() bool
	Obstruction   func() bool
}

type ElevOutputDevice struct {
	FloorIndicator     func(int)
	RequestButtonLight func(elevio.ButtonType, int, bool)
	DoorLight          func(bool)
	StopButtonLight    func(bool)
	MotorDirection     func(elevio.MotorDirection)
}

func Elevio_GetInputDevice() ElevInputDevice {
	return ElevInputDevice{
		FloorSensor:   elevio.GetFloor,
		RequestButton: elevio.GetButton,
		StopButton:    elevio.GetStop,
		Obstruction:   elevio.GetObstruction,
	}
}

func Elevio_GetOutputDevice() ElevOutputDevice {
	return ElevOutputDevice{
		FloorIndicator:     elevio.SetFloorIndicator,
		RequestButtonLight: elevio.SetButtonLamp,
		DoorLight:          elevio.SetDoorOpenLamp,
		StopButtonLight:    elevio.SetStopLamp,
		MotorDirection:     elevio.SetMotorDirection,
	}
}