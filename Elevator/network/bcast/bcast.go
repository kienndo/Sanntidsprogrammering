package bcast

import (
	"Sanntidsprogrammering/Elevator/network/conn"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"time"
	"os"
	"flag"
	localip "Sanntidsprogrammering/Elevator/network/localip"
	peers "Sanntidsprogrammering/Elevator/network/peers"
	elevio "Sanntidsprogrammering/Elevator/elevio"
	fsm "Sanntidsprogrammering/Elevator/fsm"

)

const bufSize = 1024

// Encodes received values from `chans` into type-tagged JSON, then broadcasts
// it on `port`
// Funksjonen aksepterer et variabelt antall argumenter (...) av interface{}-type
// interface{}-type er en tom interface som kan inneholde verdier av hvilken som helst type
// gyldige verdier kan være Transmitter(8080), Transmitter(8080, kanal1), Transmitter(8080, kanal1, kanal2, "en streng", 42, etObjekt)
func Transmitter(port int, chans ...interface{}) {
	checkArgs(chans...) //sjekker at argumentene er gyldige (funksjon lenger ned)
	typeNames := make([]string, len(chans)) //lager en slice av strings med lengde lik antall chans
	selectCases := make([]reflect.SelectCase, len(typeNames)) //lager en slice av reflect.SelectCase med lengde lik antall chanseler i args.
	for i, ch := range chans {
		selectCases[i] = reflect.SelectCase{ //lager en reflect.SelectCase for hver channel
			Dir:  reflect.SelectRecv, // SelectRecv : case <-Chan (mottar data fra kanal)
			Chan: reflect.ValueOf(ch), // lagrer verdien av kanalen i Chan
		}
		typeNames[i] = reflect.TypeOf(ch).Elem().String() // lagrer navnet på elementtypen til kanalen i typeNames 
	}

	conn := conn.DialBroadcastUDP(port) //lager en UDP-tilkobling som kan broadcaste til flere Receivers (fra bcast_conn_darwin.go)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port)) //lager en adresse som sender UDP-meldinger
	for {
		chosen, value, _ := reflect.Select(selectCases) //velger en case fra selectCases og lagrer verdien i value
		jsonstr, _ := json.Marshal(value.Interface()) //konverterer verdien til JSON
		ttj, _ := json.Marshal(typeTaggedJSON{ //konverterer verdien til type-tagged JSON og lagrer i ttj 
			TypeId: typeNames[chosen], //lagrer navnet på elementtypen til kanalen i TypeId
			JSON:   jsonstr, //lagrer JSON-verdien i JSON
		})
		if len(ttj) > bufSize { 
		    panic(fmt.Sprintf(
		        "Tried to send a message longer than the buffer size (length: %d, buffer size: %d)\n\t'%s'\n"+
		        "Either send smaller packets, or go to network/bcast/bcast.go and increase the buffer size",
		        len(ttj), bufSize, string(ttj)))
		}
		conn.WriteTo(ttj, addr) //sender type-tagged JSON til addr -> broadcast til alle enheter på nettverket 
    		
	}
}

// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
// Funksjonen aksepterer et variabelt antall argumenter (...) av interface{}-type
// interface{}-type er en tom interface som kan inneholde verdier av hvilken som helst type
func Receiver(port int, chans ...interface{}) {
	checkArgs(chans...)
	chansMap := make(map[string]interface{}) //lager en map med string som key og interface{} som value
	for _, ch := range chans {
		chansMap[reflect.TypeOf(ch).Elem().String()] = ch //legger elementtypen til kanalen som key og kanalen som value i chansMap. elementtype kan være en struct, int, string, etc.
	}

	var buf [bufSize]byte  //buffer for å motta data
	conn := conn.DialBroadcastUDP(port)
	for {
		n, _, e := conn.ReadFrom(buf[0:]) //leser data fra conn og lagrer i buf
		if e != nil {
			fmt.Printf("bcast.Receiver(%d, ...):ReadFrom() failed: \"%+v\"\n", port, e)
		}

		var ttj typeTaggedJSON //lager en variabel av typeTaggedJSON, som er en struct med TypeId og JSON
		json.Unmarshal(buf[0:n], &ttj) //konverterer JSON til typeTaggedJSON og lagrer i ttj, n = lengden på dataen som ble lest fra conn, buf = dataen som ble lest fra conn, &ttj = peker til ttj
		// JSON-verdier er en verdi som kan være en string, int, bool, struct, etc.
		ch, ok := chansMap[ttj.TypeId] //henter kanalen fra chansMap med key = ttj.TypeId og lagrer i ch og ok
		if !ok { //hvis kanalen ikke finnes i chansMap
			continue
		} 
		v := reflect.New(reflect.TypeOf(ch).Elem()) //lager en ny verdi av elementtypen til kanalen og lagrer i v 
		json.Unmarshal(ttj.JSON, v.Interface()) //konverterer JSON til elementtypen til kanalen og lagrer i v
		reflect.Select([]reflect.SelectCase{{ //velger en case fra selectCases
			Dir:  reflect.SelectSend, //SelectSend : case Chan <- Send (sender data til kanal)
			Chan: reflect.ValueOf(ch), //lagrer verdien av kanalen i Chan
			Send: reflect.Indirect(v), //lagrer verdien av v i Send 
		}})
	}
}

type typeTaggedJSON struct {
	TypeId string
	JSON   []byte
}

// Checks that args to Tx'er/Rx'er are valid:
//  All args must be channels
//  Element types of channels must be encodable with JSON
//  No element types are repeated
// Implementation note:
//  - Why there is no `isMarshalable()` function in encoding/json is a mystery,
//    so the tests on element type are hand-copied from `encoding/json/encode.go`
func checkArgs(chans ...interface{}) {
	n := 0
	for range chans {
		n++
	}
	elemTypes := make([]reflect.Type, n) //lager en slice av reflect.Type med lengde lik antall chans

	for i, ch := range chans {
		// Must be a channel
		if reflect.ValueOf(ch).Kind() != reflect.Chan {
			panic(fmt.Sprintf(
				"Argument must be a channel, got '%s' instead (arg# %d)",
				reflect.TypeOf(ch).String(), i+1))
		}

		elemType := reflect.TypeOf(ch).Elem()

		// Element type must not be repeated
		for j, e := range elemTypes {
			if e == elemType {
				panic(fmt.Sprintf(
					"All channels must have mutually different element types, arg# %d and arg# %d both have element type '%s'",
					j+1, i+1, e.String()))
			}
		}
		elemTypes[i] = elemType

		// Element type must be encodable with JSON
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

func RunBroadcast() {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer") //flag.StringVar lagrer verdien av id i id variabelen 
	flag.Parse() 

	if id == "" { //hvis id er tom
		localIP, err := localip.LocalIP() //henter IP-adressen til enheten
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid()) //id = peer-(IP-adresse)-(prosess-ID), os.Getpid() = henter prosess-ID, 
	}

	// Denne kanalen mottar oppdateringer om hvilke peers som er tilkoblet nettverket, inneholder Peers []string, New string og Lost []string
	peerUpdateCh := make(chan peers.PeerUpdate)

	// Denne kanalen aktiverer eller deaktiverer transmitteren
	peerTxEnable := make(chan bool) 
	go peers.Transmitter(15647, id, peerTxEnable) //Sender id til alle enheter på nettverket via UDP (fra peers.go)
	go peers.Receiver(15647, peerUpdateCh) //Mottar data fra alle enheter på nettverket via UDP og sender oppdateringer til peerUpdateCh (fra peers.go)

	helloTx := make(chan elevio.Elevator) //kanal for å sende HelloMsg -> sjekker argumenter og lagrer i selectCases og typeNames
	helloRx := make(chan elevio.Elevator) //kanal for å motta HelloMsg -> sjekker argumenter og lagrer i chansMap 

	go Transmitter(16569, helloTx) //Sender HelloMsg til alle enheter på nettverket via UDP (fra bcast.go)
	go Receiver(16569, helloRx) // Mottar HelloMsg fra alle enheter på nettverket via UDP (fra bcast.go)

	go func() {
		for {
			ElevatorMessage := fsm.RunningElevator 
			helloTx <- ElevatorMessage //Sender helloMsg til helloTx-kanalen hvert sekund
			time.Sleep(1 * time.Second) //Venter 1 sekund
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh: //hvis det kommer en melding på peerUpdateCh-kanalen
			fmt.Printf("Peer update:\n") //skriver ut "Peer update:"
			fmt.Printf("  Peers:    %q\n", p.Peers) //skriver ut "Peers: " og innholdet i p.Peers
			fmt.Printf("  New:      %q\n", p.New) //skriver ut "New: " og innholdet i p.New
			fmt.Printf("  Lost:     %q\n", p.Lost) //skriver ut "Lost: " og innholdet i p.Lost

		case a := <-helloRx: //hvis det kommer en melding på helloRx-kanalen
			fmt.Printf("Received: %#v\n", a) //skriver ut "Received: " og innholdet i a
		}
	}
}