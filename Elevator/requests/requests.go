package requests

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
)

func ShouldClearImmediately(e elevio.Elevator, btn_floor int, btn_type elevio.ButtonType) int {
	switch e.Config.ClearRequestVariant {
	case elevio.CV_All:
		if e.Floor == btn_floor {
			return 1
		}
		return 0
	case elevio.CV_InDirn:
		if e.Floor == btn_floor &&
			(e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp) ||
			(e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown) ||
			e.Dirn == elevio.MD_Stop ||
			btn_type == elevio.BT_Cab {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func ChooseDirection(e elevio.Elevator) elevio.DirnBehaviourPair {
	switch e.Dirn {
	case elevio.MD_Up:
		if IfFloorAbove(e) != 0 {
			return elevio.DirnBehaviourPair{elevio.MD_Up, elevio.EB_Moving}
		}
		if IfFloorHere(e) != 0 {
			return elevio.DirnBehaviourPair{elevio.MD_Down, elevio.EB_DoorOpen}
		}
		if IfFloorBelow(e) != 0 {
			return elevio.DirnBehaviourPair{elevio.MD_Down, elevio.EB_Moving}
		}
		return elevio.DirnBehaviourPair{elevio.MD_Stop, elevio.EB_Idle}
	case elevio.MD_Down:
		if IfFloorBelow(e) != 0 {
			return elevio.DirnBehaviourPair{elevio.MD_Down, elevio.EB_Moving}
		}
		if IfFloorHere(e) != 0 {
			return elevio.DirnBehaviourPair{elevio.MD_Up, elevio.EB_DoorOpen}
		}
		if IfFloorAbove(e) != 0 {
			return elevio.DirnBehaviourPair{elevio.MD_Up, elevio.EB_Moving}
		}
		return elevio.DirnBehaviourPair{elevio.MD_Stop, elevio.EB_Idle}
	case elevio.MD_Stop:

		if IfFloorHere(e) != 0 {
			return elevio.DirnBehaviourPair{elevio.MD_Stop, elevio.EB_DoorOpen}
		}
		if IfFloorAbove(e) != 0 {
			return elevio.DirnBehaviourPair{elevio.MD_Up, elevio.EB_Moving}
		}
		if IfFloorBelow(e) != 0 {
			return elevio.DirnBehaviourPair{elevio.MD_Down, elevio.EB_Moving}
		}
		return elevio.DirnBehaviourPair{elevio.MD_Stop, elevio.EB_Idle}
	default:
		return elevio.DirnBehaviourPair{elevio.MD_Stop, elevio.EB_Idle}
	}
}

func IfFloorAbove(e elevio.Elevator) int {
	for f := e.Floor + 1; f < elevio.N_FLOORS; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if e.Request[f][btn] != false {
				return 1
			}
		}
	}
	return 0
}

func IfFloorBelow(e elevio.Elevator) int {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if e.Request[f][btn] != false {
				return 1
			}
		}
	}
	return 0
}

func IfFloorHere(e elevio.Elevator) int {
	for btn := 0; btn < elevio.N_BUTTONS; btn++ {
		if e.Request[e.Floor][btn] != false {
			return 1
		}
	}
	return 0
}

func ClearAtCurrentFloor(e elevio.Elevator) elevio.Elevator {

	e.Request[e.Floor][elevio.BT_Cab] = false

	switch e.Dirn {
	case elevio.MD_Up:
		if (IfFloorAbove(e) == 0) && (e.Request[e.Floor][elevio.BT_HallUp] == false) {
			e.Request[e.Floor][elevio.BT_HallDown] = false
		}
		e.Request[e.Floor][elevio.BT_HallUp] = false
	case elevio.MD_Down:
		if (IfFloorBelow(e) == 0) && (e.Request[e.Floor][elevio.BT_HallDown] == false) {
			e.Request[e.Floor][elevio.BT_HallUp] = false
		}
		e.Request[e.Floor][elevio.BT_HallDown] = false
	default:
		e.Request[e.Floor][elevio.BT_HallUp] = false
		e.Request[e.Floor][elevio.BT_HallDown] = false
	}
	return e
}

func ShouldStop(e elevio.Elevator) int {
	switch e.Dirn {
	case elevio.MD_Down:
		if e.Request[e.Floor][elevio.BT_HallDown] != false {
			return 1
		}
		if e.Request[e.Floor][elevio.BT_Cab] != false {
			return 1
		}
		if IfFloorBelow(e) == 0 {
			return 1
		}
		return 0
	case elevio.MD_Up:
		if e.Request[e.Floor][elevio.BT_HallUp] != false {
			return 1
		}
		if e.Request[e.Floor][elevio.BT_Cab] != false {
			return 1
		}
		if IfFloorAbove(e) == 0 {
			return 1
		}
		return 0
	default:
		return 1
	}
}

/*func MergeHallAndCab(hallRequests [elevio.N_FLOORS][2]bool, cabRequests [elevio.N_FLOORS]bool) [elevio.N_FLOORS][elevio.N_BUTTONS]bool {
	var requests [elevio.N_FLOORS][elevio.N_BUTTONS]bool
	for i := range requests {
		requests[i] = [elevio.N_BUTTONS]bool{hallRequests[i][0], hallRequests[i][1], cabRequests[i]}
	}
	return requests
}

func GetExecutedHallOrder(e elevio.Elevator) elevio.ButtonEvent {
	requests := MergeHallAndCab(e.HallRequests, e.CabRequests)
	switch e.Dirn {
	case elevio.MD_Stop:
		if requests[e.Floor][elevio.BT_HallUp] && IfFloorAbove(e) == false {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp}
		} else if requests[e.Floor][elevio.BT_HallDown] && !IfFloorBelow(e) {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown}
		} else if requests[e.Floor][elevio.BT_HallDown] && IfFloorBelow(e) {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown}
		} else {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp}
		}
	case elevio.MD_Up:
		if !requests[e.Floor][elevio.BT_HallUp] && !IfFloorAbove(e) {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown}
		} else {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp}
		}
	case elevio.MD_Down:
		if !requests[e.Floor][elevio.BT_HallDown] && !IfFloorBelow(e) {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallUp}
		} else {
			return elevio.ButtonEvent{Floor: e.Floor, Button: elevio.BT_HallDown}
		}
	default:
		panic("Elevator request error")
	}
}*/