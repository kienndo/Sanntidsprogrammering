package bcast

import (
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"
	conn "Sanntidsprogrammering/Elevator/network/conn"
	localip "Sanntidsprogrammering/Elevator/network/localip"
	peers "Sanntidsprogrammering/Elevator/network/peers"
	requests "Sanntidsprogrammering/Elevator/requests"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"reflect"
	"time"

)

var( 
	ID string
)

const bufSize = 1024

func Transmitter(port int, chans ...interface{}) {
	checkArgs(chans...)
	typeNames := make([]string, len(chans))
	selectCases := make([]reflect.SelectCase, len(typeNames)) 
	for i, ch := range chans {
		selectCases[i] = reflect.SelectCase{ 
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch), 
		}
		typeNames[i] = reflect.TypeOf(ch).Elem().String() 
	}
	conn := conn.DialBroadcastUDP(port) 
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("10.100.23.255:%d", port)) 
	for {
		chosen, value, _ := reflect.Select(selectCases) 
		jsonstr, _ := json.Marshal(value.Interface()) 
		ttj, _ := json.Marshal(typeTaggedJSON{ 
			TypeId: typeNames[chosen], 
			JSON:   jsonstr, 
		})
		if len(ttj) > bufSize { 
		    panic(fmt.Sprintf(
		        "Tried to send a message longer than the buffer size (length: %d, buffer size: %d)\n\t'%s'\n"+
		        "Either send smaller packets, or go to network/bcast/bcast.go and increase the buffer size",
		        len(ttj), bufSize, string(ttj)))
		}
		conn.WriteTo(ttj, addr) 
    		
	}
}

func Receiver(port int, chans ...interface{}) {
	checkArgs(chans...)
	chansMap := make(map[string]interface{}) 
	for _, ch := range chans {
		chansMap[reflect.TypeOf(ch).Elem().String()] = ch 
	}

	var buf [bufSize]byte 
	conn := conn.DialBroadcastUDP(port)
	for {
		n, _, e := conn.ReadFrom(buf[0:])
		if e != nil {
			fmt.Printf("bcast.Receiver(%d, ...):ReadFrom() failed: \"%+v\"\n", port, e)
		}

		var ttj typeTaggedJSON
		json.Unmarshal(buf[0:n], &ttj) 

		ch, ok := chansMap[ttj.TypeId] 
		if !ok { 
			continue
		} 
		v := reflect.New(reflect.TypeOf(ch).Elem()) 
		json.Unmarshal(ttj.JSON, v.Interface())
		reflect.Select([]reflect.SelectCase{{ 
			Dir:  reflect.SelectSend, 
			Chan: reflect.ValueOf(ch), 
			Send: reflect.Indirect(v), 
		}})

		time.Sleep(3)
	}
}

type typeTaggedJSON struct {
	TypeId string
	JSON   []byte
}

func checkArgs(chans ...interface{}){
	n := 0
	for range chans {
		n++
	}
	elemTypes := make([]reflect.Type, n)

	for i, ch := range chans {
		
		if reflect.ValueOf(ch).Kind() != reflect.Chan {
			panic(fmt.Sprintf(
				"Argument must be a channel, got '%s' instead (arg# %d)",
				reflect.TypeOf(ch).String(), i+1))
		}

		elemType := reflect.TypeOf(ch).Elem()

		for j, e := range elemTypes {
			if e == elemType {
				panic(fmt.Sprintf(
					"All channels must have mutually different element types, arg# %d and arg# %d both have element type '%s'",
					j+1, i+1, e.String()))
			}
		}
		elemTypes[i] = elemType

		checkTypeRecursive(elemType, []int{i+1})
	}
}

func checkTypeRecursive(val reflect.Type, offsets []int){
	switch val.Kind() {
	case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		panic(fmt.Sprintf(
			"Channel element type must be supported by JSON, got '%s' instead (nested arg# %v)",
			val.String(), offsets))
	case reflect.Map:
		if val.Key().Kind() != reflect.String {
			panic(fmt.Sprintf(
				"Channel element type must be supported by JSON, got '%s' instead (map keys must be 'string') (nested arg# %v)",
				val.String(), offsets))
		}
		checkTypeRecursive(val.Elem(), offsets)
	case reflect.Array, reflect.Ptr, reflect.Slice:
		checkTypeRecursive(val.Elem(), offsets)
	case reflect.Struct:
		for idx := 0; idx < val.NumField(); idx++ {
			checkTypeRecursive(val.Field(idx).Type, append(offsets, idx+1))
		}
	}
}

func RunBroadcast(ElevatorMessageTX chan elevio.Elevator, addr int) {
	flag.StringVar(&ID, "id", "", "id of this peer") 
	flag.Parse() 

	if ID == "" { 
		localIP, err := localip.LocalIP() 
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		ID = fmt.Sprintf("%s:%d", localIP, os.Getpid())
	}

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool) 

	go peers.Transmitter(15646, ID, peerTxEnable)
	go peers.Receiver(15646, peerUpdateCh)

	ElevatorMessageRX := make(chan elevio.Elevator)

	go Transmitter(addr, ElevatorMessageTX) 
	go Receiver(addr, ElevatorMessageRX)
	
	go func() {
		for {

			ElevatorMessage := fsm.RunningElevator
			ElevatorMessageTX <- ElevatorMessage
			
			for i:= 0; i<elevio.N_FLOORS; i++{
				for j:= 0; j<2; j++{
					requests.RequestMutex.Lock()
					fsm.RunningElevator.HallRequests[i][j] = false
					
					requests.RequestMutex.Unlock()
				}
			}
			
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		select {
		case p := <-peerUpdateCh: 
			fmt.Printf("Slave update:\n") 
			fmt.Printf("  Slaves:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New) 
			fmt.Printf("  Lost:     %q\n", p.Lost) 

		case a := <-ElevatorMessageRX: 
			fmt.Printf("Received: %#v\n", a)
		}
	}
}
