package requests

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
)


// This function
func ShouldClearImmediately(e elevio.Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	switch e.Config.ClearRequestVariant {
	case elevio.CV_All:
		return e.Floor == btn_floor
	case elevio.CV_InDirn:
		return e.Floor == btn_floor && ((e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp) ||
			(e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown) ||
			(e.Dirn == elevio.MD_Stop) || (btn_type == elevio.BT_Cab))
	default:
		return false
	}
}

func ChooseDirection(e elevio.Elevator) elevio.DirnBehaviourPair {
	switch e.Dirn {
	case elevio.MD_Up:
		if IfFloorAbove(e) {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevio.EB_Moving}
		} else if IfFloorHere(e) {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevio.EB_DoorOpen}
		} else if IfFloorBelow(e) {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevio.EB_Moving}
		} else {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevio.EB_Idle}
		}
		case elevio.MD_Down:
		if IfFloorBelow(e) {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevio.EB_Moving}
		} else if IfFloorHere(e) {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevio.EB_DoorOpen}
		} else if IfFloorAbove(e) {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevio.EB_Moving}
		} else {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevio.EB_Idle}
		}
	case elevio.MD_Stop:
		if IfFloorHere(e) {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevio.EB_DoorOpen}
		} else if IfFloorAbove(e) {
    		return elevio.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevio.EB_Moving}
		} else if IfFloorBelow(e) {
    		return elevio.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevio.EB_Moving}
		} else {
    		return elevio.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevio.EB_Idle}
		}
		default:
    		return elevio.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevio.EB_Idle}
		}

}

func IfFloorAbove(e elevio.Elevator) bool {
	for f := e.Floor + 1; f < elevio.N_FLOORS; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if e.Request[f][btn] == 1 {
				return true
			}
		}
	}
	return false
}

func IfFloorBelow(e elevio.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if e.Request[f][btn] == 1 {
				return true
			}
		}
	}
	return false
}

func IfFloorHere(e elevio.Elevator) bool {
	for btn := 0; btn < elevio.N_BUTTONS; btn++ {
		if e.Request[e.Floor][btn] == 1 {
			return true
		}
	}
	return false
}
func ClearAtCurrentFloor(e elevio.Elevator) elevio.Elevator {
	switch e.Config.ClearRequestVariant {
	case elevio.CV_All:
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			e.Request[e.Floor][btn] = 0
		}
	case elevio.CV_InDirn:
		e.Request[e.Floor][elevio.BT_Cab] = 0
		switch e.Dirn {
		case elevio.MD_Up:
			if !IfFloorAbove(e) && e.Request[e.Floor][elevio.BT_HallUp] == 0 {
				e.Request[e.Floor][elevio.BT_HallDown] = 0
			}
			e.Request[e.Floor][elevio.BT_HallUp] = 0
		case elevio.MD_Down:
			if !IfFloorBelow(e) && e.Request[e.Floor][elevio.BT_HallDown] == 0 {
				e.Request[e.Floor][elevio.BT_HallUp] = 0
			}
			e.Request[e.Floor][elevio.BT_HallDown] = 0
		case elevio.MD_Stop:
			fallthrough
		default:
			e.Request[e.Floor][elevio.BT_HallUp] = 0
			e.Request[e.Floor][elevio.BT_HallDown] = 0
		}
	default:
	}
	return e
}

func ShouldStop(e elevio.Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		return e.Request[e.Floor][elevio.BT_HallDown] == 1 || e.Request[e.Floor][elevio.BT_Cab] == 1 || !IfFloorBelow(e)
	case elevio.MD_Up:
		return e.Request[e.Floor][elevio.BT_HallUp] == 1 || e.Request[e.Floor][elevio.BT_Cab] == 1 || !IfFloorAbove(e)
	case elevio.MD_Stop:
		fallthrough
	default:
		return true
	}
}