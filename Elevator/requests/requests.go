package requests

// Direct translation from C to Golang, retrieved from https://github.com/TTK4145/Project-resources/tree/master/elev_algo

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	"sync"
)

var FsmMutex = sync.Mutex{}

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
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevio.EB_Moving}
		}
		if IfFloorHere(e) != 0 {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevio.EB_DoorOpen}
		}
		if IfFloorBelow(e) != 0 {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevio.EB_Moving}
		}
		return elevio.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevio.EB_Idle}
	case elevio.MD_Down:
		if IfFloorBelow(e) != 0 {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevio.EB_Moving}
		}
		if IfFloorHere(e) != 0 {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevio.EB_DoorOpen}
		}
		if IfFloorAbove(e) != 0 {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevio.EB_Moving}
		}
		return elevio.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevio.EB_Idle}
	case elevio.MD_Stop:
		if IfFloorHere(e) != 0 {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevio.EB_DoorOpen}
		}
		if IfFloorAbove(e) != 0 {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevio.EB_Moving}
		}
		if IfFloorBelow(e) != 0 {
			return elevio.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevio.EB_Moving}
		}
		return elevio.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevio.EB_Idle}
	default:
		return elevio.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevio.EB_Idle}
	}
}

func IfFloorAbove(e elevio.Elevator) int {
	for f := e.Floor + 1; f < elevio.N_FLOORS; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if e.Request[f][btn] {
				return 1
			}
		}
	}
	return 0
}

func IfFloorBelow(e elevio.Elevator) int {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if e.Request[f][btn] {
				return 1
			}
		}
	}
	return 0
}

func IfFloorHere(e elevio.Elevator) int {
	for btn := 0; btn < elevio.N_BUTTONS; btn++ {
		if e.Request[e.Floor][btn] {
			return 1
		}
	}
	return 0
}

func ClearAtCurrentFloor(e elevio.Elevator) elevio.Elevator {
	FsmMutex.Lock()
	FsmMutex.Unlock()
	e.Request[e.Floor][elevio.BT_Cab] = false

	switch e.Dirn {
	case elevio.MD_Up:
		if (IfFloorAbove(e) == 0) && (!e.Request[e.Floor][elevio.BT_HallUp]) {
			e.Request[e.Floor][elevio.BT_HallDown] = false
			e.HallRequests[e.Floor][elevio.BT_HallDown] = false
		}
		e.Request[e.Floor][elevio.BT_HallUp] = false
		e.HallRequests[e.Floor][elevio.BT_HallUp] = false
	case elevio.MD_Down:
		if (IfFloorBelow(e) == 0) && (!e.Request[e.Floor][elevio.BT_HallDown]) {
			e.Request[e.Floor][elevio.BT_HallUp] = false
			e.HallRequests[e.Floor][elevio.BT_HallUp] = false
		}
		e.Request[e.Floor][elevio.BT_HallDown] = false
		e.HallRequests[e.Floor][elevio.BT_HallDown] = false
	default:
		e.Request[e.Floor][elevio.BT_HallUp] = false
		e.Request[e.Floor][elevio.BT_HallDown] = false
		e.HallRequests[e.Floor][elevio.BT_HallUp] = false
		e.HallRequests[e.Floor][elevio.BT_HallDown] = false
	}
	return e
}

func ShouldStop(e elevio.Elevator) int {
	FsmMutex.Lock()
	switch e.Dirn {
	case elevio.MD_Down:
		if e.Request[e.Floor][elevio.BT_HallDown] {
			return 1
		}
		if e.Request[e.Floor][elevio.BT_Cab] {
			return 1
		}
		if IfFloorBelow(e) == 0 {
			return 1
		}
		return 0
	case elevio.MD_Up:
		if e.Request[e.Floor][elevio.BT_HallUp] {
			return 1
		}
		if e.Request[e.Floor][elevio.BT_Cab] {
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
