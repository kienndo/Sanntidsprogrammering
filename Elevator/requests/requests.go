package requests

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
)


<<<<<<< HEAD

=======
func requests_mergeHallAndCab(hallRequests [elevio.N_FLOORS][2]bool, cabRequests [elevio.N_FLOORS]bool) [elevio.N_FLOORS][elevio.N_BUTTONS]bool {
	var requests [elevio.N_FLOORS][elevio.N_BUTTONS]bool
	for i := range requests {
		requests[i] = [elevio.N_BUTTONS]bool{hallRequests[i][0], hallRequests[i][1], cabRequests[i]}
	}
	return requests
}

// This function
>>>>>>> 0cb603f (8.mars)
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
	Request := requests_mergeHallAndCab(e.HallRequests, e.CabRequests)
	for f := e.Floor + 1; f < elevio.N_FLOORS; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if Request[f][btn] {
				return true
			}
		}
	}
	return false
}

func IfFloorBelow(e elevio.Elevator) bool {
	Request := requests_mergeHallAndCab(e.HallRequests, e.CabRequests)
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if Request[f][btn] {
				return true
			}
		}
	}
	return false
}

func IfFloorHere(e elevio.Elevator) bool {
	Request := requests_mergeHallAndCab(e.HallRequests, e.CabRequests)
	for btn := 0; btn < elevio.N_BUTTONS; btn++ {
		if Request[e.Floor][btn] {
			return true
		}
	}
	return false
}
func ClearAtCurrentFloor(e elevio.Elevator) elevio.Elevator {
	Request := requests_mergeHallAndCab(e.HallRequests, e.CabRequests)
	switch e.Config.ClearRequestVariant {
	case elevio.CV_All:
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			Request[e.Floor][btn] = false
		}
	case elevio.CV_InDirn:
		Request[e.Floor][elevio.BT_Cab] = false
		switch e.Dirn {
		case elevio.MD_Up:
			if !IfFloorAbove(e) && Request[e.Floor][elevio.BT_HallUp] == false {
				Request[e.Floor][elevio.BT_HallDown] = false
			}
			Request[e.Floor][elevio.BT_HallUp] = false
		case elevio.MD_Down:
			if !IfFloorBelow(e) && Request[e.Floor][elevio.BT_HallDown] == false {
				Request[e.Floor][elevio.BT_HallUp] = false
			}
			Request[e.Floor][elevio.BT_HallDown] = false
		case elevio.MD_Stop:
			fallthrough
		default:
			Request[e.Floor][elevio.BT_HallUp] = false
			Request[e.Floor][elevio.BT_HallDown] = false
		}
	default:
	}
	return e
}

func ShouldStop(e elevio.Elevator) bool {
	Request := requests_mergeHallAndCab(e.HallRequests, e.CabRequests)
	switch e.Dirn {
	case elevio.MD_Down:
		return Request[e.Floor][elevio.BT_HallDown] == true || Request[e.Floor][elevio.BT_Cab] == true || !IfFloorBelow(e)
	case elevio.MD_Up:
		return Request[e.Floor][elevio.BT_HallUp] == true || Request[e.Floor][elevio.BT_Cab] == true || !IfFloorAbove(e)
	case elevio.MD_Stop:
		fallthrough
	default:
		return true
	}
}

func requests_getExecutedHallOrder(e elevio.Elevator) elevio.ButtonEvent {
	requests := requests_mergeHallAndCab(e.HallRequests, e.CabRequests)
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
}