package elevio

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Constants for the number of floors and buttons
const (
	N_FLOORS	= 4
	N_BUTTONS	= 3
	_pollRate = 20 * time.Millisecond
)

// Different used variables
var _initialized bool = false
var _numFloors int = 4
var _mtx sync.Mutex
var _conn net.Conn

// Struct for the elevator
type Elevator struct {
	Floor int // Which floor
	Dirn MotorDirection // Which direction the elevator is going in
	Behaviour ElevatorBehaviour // Which state the elevator is in
	Request [N_FLOORS][N_BUTTONS]int // Which request the elevator has
	Config Config // Configuration of the elevator
}

// Struct for the configuration of the elevator
type Config struct {
	ClearRequestVariant ClearRequestVariant // I think the other code defines how it is initialized
	DoorOpenDuration float64 // How long the door is open
}

//clear-funksjon på om alt skal fjernes eller bare noe
type ClearRequestVariant int
const (
	CV_All ClearRequestVariant = 0 //If i have not understood this wrong, this is the variant for clearing all requests
	CV_InDirn = 1 // Only clearing in one direction
)

// A struct with a pair for which way the elevator is going and what kind of state it is in
type DirnBehaviourPair struct {
	Dirn MotorDirection
	Behaviour ElevatorBehaviour
}


// Enum for the direction of the elevator
type MotorDirection int
const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

// Enum for the different types of buttons
type ButtonType int
const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

//enum for states for the FSM
type ElevatorBehaviour int
const (
	EB_Idle     ElevatorBehaviour = 0
	EB_DoorOpen                   = 1
	EB_Moving                     = 2
)

// Struct for the different types of button events including floor and button
type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

// Initialized the elevator through talking to the given address and choosing the number of floors
func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

// Function that will need to be used when the elevator starts getting orders:
// Setting motor direction
func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

// Setting button lamp
func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

// Setting the floor indicator (shows which floor we are on)
func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

// Setting the door open lamp
func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

// Setting the stop lamp
func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

// Checks if any of the buttons has been pressed
func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- ButtonEvent{f, ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

// Checks if the floor sensor registers any floors through "GetFloor()"
func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

// Checks if the stop button has been pressed
func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

// Checks if the obstruction switch has been turned on
func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

// Function that returns if a button on a specific floor has been pushed (bool)
func GetButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

// Function that returns the floor the elevator is currently on
func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

// Function that returns if the stop button has been pushed (bool)
func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

// Function that returns if the obstruction switch has been turned on (bool)
func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}


// Reads data from a connection (with mutex)
func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

// Writes data to a connection (with mutex)
func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

// Converts bool to byte
func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

// Converts byte to bool
func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}

func EbToString(eb ElevatorBehaviour) string {
	switch eb {
	case EB_Idle:
		return "EB_Idle"
	case EB_DoorOpen:
		return "EB_DoorOpen"
	case EB_Moving:
		return "EB_Moving"
	default:
		return "EB_UNDEFINED"
	}
}

//Functions for converting the different enums to strings
func ElevioDirnToString(d MotorDirection) string {
	switch d {
	case MD_Up:
		return "D_Up"
	case MD_Down:
		return "D_Down"
	case MD_Stop:
		return "D_Stop"
	default:
		return "D_UNDEFINED"
	}
}

func ElevioButtonToString(b ButtonType) string {
	switch b {
	case BT_HallUp:
		return "B_HallUp"
	case BT_HallDown:
		return "B_HallDown"
	case BT_Cab:
		return "B_Cab"
	default:
		return "B_UNDEFINED"
	}
}
