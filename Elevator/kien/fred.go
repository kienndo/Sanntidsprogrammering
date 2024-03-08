package kien

import (
	//fsm 	"Sanntidsprogrammering/Elevator/fsm"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	"time"
	requests "Sanntidsprogrammering/Elevator/requests"
)

type Elevator struct {
	Floor            int
	Dirn             elevio.MotorDirection
	Behaviour        elevio.ElevatorBehaviour
	CabRequests      [elevio.N_FLOORS]bool
	HallRequests     [elevio.N_FLOORS][2]bool
	doorOpenDuration time.Duration
}


func FSM(
	floorCh <-chan int,
	obstrCh <-chan bool,
	cabButtonEventCh <-chan elevio.ButtonEvent,
	initCabRequestsCh <-chan [elevio.N_FLOORS]bool,
	assignedOrdersCh <-chan [elevio.N_FLOORS][2]bool,
	executedHallOrderCh chan<- elevio.ButtonEvent,
	localElevDataCh chan<- elevio.Elevator,
	isAliveCh chan<- bool) {

	var (
		
		//elevio.Config.DoorOpenDuration: 3 * time.Second
		executedHallOrder      = elevio.ButtonEvent{}
		executedHallOrderTimer = time.NewTimer(time.Hour)

		obstr         = false
		doorOpenTimer = time.NewTimer(time.Hour)
		elevDataTimer = time.NewTimer(time.Hour)

		e = elevio.Elevator{
			Floor:       -1,
			Dirn:        elevio.MD_Stop,
			Behaviour:   elevio.EB_Idle,
			CabRequests: [elevio.N_FLOORS]bool{false, false, false, false},
			HallRequests: [elevio.N_FLOORS][2]bool{
				{false, false},
				{false, false},
				{false, false},
				{false, false}},
			
		}
	)
	doorOpenTimer.Stop()
	elevDataTimer.Stop()
	executedHallOrderTimer.Stop()

initialization:
	for {
		select {
		case e.Floor = <-floorCh:
			elevio.SetMotorDirection(elevio.MD_Stop)

			e.CabRequests = <-initCabRequestsCh
			 for floor, order := range e.CabRequests {
			 	elevio.SetButtonLamp(elevio.BT_Cab, floor, order)
			 }
			var tempDirnBehav elevio.DirnBehaviourPair = requests.ChooseDirection(e)
			e.Dirn = tempDirnBehav.Dirn
			e.Behaviour = tempDirnBehav.Behaviour
			elevDataTimer.Reset(1)

			switch e.Behaviour {
			case elevio.EB_Idle:
			case elevio.EB_DoorOpen:
				elevio.SetDoorOpenLamp(true)
				doorOpenTimer.Reset(time.Duration(e.Config.DoorOpenDuration))

			case elevio.EB_Moving:
				elevio.SetMotorDirection(e.Dirn)
			}
			break initialization
		default:
			elevio.SetMotorDirection(elevio.MD_Down)
			time.Sleep(10 * time.Millisecond)
		}
	}

	for {
		select {
		case e.HallRequests = <-assignedOrdersCh:
			switch e.Behaviour {
			case elevio.EB_DoorOpen:
			case elevio.EB_Moving:
			case elevio.EB_Idle:
				var tempDirnBehav elevio.DirnBehaviourPair = requests.ChooseDirection(e)
				e.Dirn = tempDirnBehav.Dirn
				e.Behaviour = tempDirnBehav.Behaviour
				elevDataTimer.Reset(1)

				switch e.Behaviour {
				case elevio.EB_Idle:
				case elevio.EB_DoorOpen:
					elevio.SetDoorOpenLamp(true)
					doorOpenTimer.Reset(time.Duration(e.Config.DoorOpenDuration))

				case elevio.EB_Moving:
					elevio.SetMotorDirection(e.Dirn)
				}
			}
		case cabButtonEvent := <-cabButtonEventCh:
			e.CabRequests[cabButtonEvent.Floor] = true
			elevio.SetButtonLamp(elevio.BT_Cab, cabButtonEvent.Floor, true)
			elevDataTimer.Reset(1)
			switch e.Behaviour {
			case elevio.EB_DoorOpen:
			case elevio.EB_Moving:
			case elevio.EB_Idle:
				var tempDirnBehav elevio.DirnBehaviourPair = requests.ChooseDirection(e)
				e.Dirn = tempDirnBehav.Dirn
				e.Behaviour = tempDirnBehav.Behaviour
				switch e.Behaviour {
				case elevio.EB_Idle:
				case elevio.EB_DoorOpen:
					elevio.SetDoorOpenLamp(true)
					doorOpenTimer.Reset(time.Duration(e.Config.DoorOpenDuration))

				case elevio.EB_Moving:
					elevio.SetMotorDirection(e.Dirn)
				}
			}
		case e.Floor = <-floorCh:
			elevio.SetFloorIndicator(e.Floor)
			switch e.Behaviour {
			case elevio.EB_Idle:
			case elevio.EB_DoorOpen:
			case elevio.EB_Moving:
				elevDataTimer.Reset(1)
				if requests.ShouldStop(e) == true {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevio.SetDoorOpenLamp(true)
					e.Behaviour = elevio.EB_DoorOpen
					if !obstr {
						doorOpenTimer.Reset(time.Duration(e.Config.DoorOpenDuration))
					}
				}
			}
		case <-doorOpenTimer.C:
			switch e.Behaviour {
			case elevio.EB_Idle:
			case elevio.EB_Moving:
			case elevio.EB_DoorOpen:
				e.CabRequests[e.Floor] = false
				elevio.SetButtonLamp(elevio.BT_Cab, e.Floor, false)
				executedHallOrder = requests.GetExecutedHallOrder(e) 
				e.HallRequests[executedHallOrder.Floor][executedHallOrder.Button] = false
				executedHallOrderTimer.Reset(1)
				var tempDirnBehav elevio.DirnBehaviourPair = requests.ChooseDirection(e)
				e.Dirn = tempDirnBehav.Dirn
				e.Behaviour = tempDirnBehav.Behaviour
				elevDataTimer.Reset(1)
				switch e.Behaviour {
				case elevio.EB_DoorOpen:
					doorOpenTimer.Reset(time.Duration(e.Config.DoorOpenDuration))
				case elevio.EB_Moving, elevio.EB_Idle:
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(e.Dirn)
				}
			}
		case obstr = <-obstrCh:
			if obstr {
				doorOpenTimer.Stop()
			}
			switch e.Behaviour {
			case elevio.EB_Idle:
			case elevio.EB_Moving:
			case elevio.EB_DoorOpen:
				if !obstr {
					doorOpenTimer.Reset(time.Duration(e.Config.DoorOpenDuration))
				}
			}
		case <-elevDataTimer.C:
				elevDataTimer.Reset(1)
		case <-executedHallOrderTimer.C:
			select {
			case executedHallOrderCh <- executedHallOrder:
			default:
				executedHallOrderTimer.Reset(1)
			}
		}
	}
}

// func getElevatorData(e Elevator) elevio.Elevator {
// 	dirnToString := map[elevio.MotorDirection]string{
// 		elevio.MD_Down: "down",
// 		elevio.MD_Up:   "up",
// 		elevio.MD_Stop: "stop"}

// 	return elevio.Elevator{
// 		Behavior:    string(e.Behaviour),
// 		Floor:       e.Floor,
// 		Direction:   dirnToString[e.Dirn],
// 		CabRequests: e.CabRequests}
// }

