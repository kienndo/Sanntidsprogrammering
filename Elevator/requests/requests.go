package requests

import (
	elevio "Sanntidsprogrammering/elevio"
)


//
func requests_shouldClearImmediately(e elevio.Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
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