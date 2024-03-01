package elevio

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _numFloors int = 4
var _mtx sync.Mutex
var _conn net.Conn

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
