package peers

import (
	"Sanntidsprogrammering/Elevator/network/conn"
	"fmt"
	"net"
	"sort"
	"time"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

const interval = 15 * time.Millisecond
const timeout = 500 * time.Millisecond

// Transmitter sender data over nettverket via UDP
// transmitEnable = Receive-only channel som kan aktivere eller deaktivere transmitteren

func Transmitter(port int, id string, transmitEnable <-chan bool) {

	conn := conn.DialBroadcastUDP(port) //UDP tilkobling som kan broadcaste til flere Receivers (fra bcast_conn_darwin.go)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port)) //lager en adresse som sender UDP-meldinger
	// 255.255.255.255 = IP-adresse for å broadcaste til alle enheter på nettverket
	// port = porten som meldingen skal sendes til
	enable := true
	for {
		select {
		case enable = <-transmitEnable: //hvis det kommer en melding på transmitEnable-kanalen, settes enable=true/false
		case <-time.After(interval): // Etter interval = 15 * time.Millisecond = 0.015 sekunder
		}
		if enable {
			conn.WriteTo([]byte(id), addr) //sender id (string) til addr -> broadcast til alle enheter på nettverket
		}
	}
}
// Receiver mottar data over nettverket via UDP og sender oppdateringer til peerUpdateChan
// port = porten som meldinger skal mottas på
// peerUpdateCh chan<- PeerUpdate = Send-only channel som sender oppdateringer 
// PeerUpdate er en struct som inneholder Peers []string, New string og Lost []string

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {

	var buf [1024]byte //buffer for å motta data
	var p PeerUpdate 
	lastSeen := make(map[string]time.Time) //map som inneholder id og tidspunkt for når id ble sist sett

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval)) //setter en deadline for å motta data
		n, _, _ := conn.ReadFrom(buf[0:]) //leser data fra conn og lagrer i buf

		id := string(buf[:n]) //id = data som ble lest fra conn 

		// Adding new connection
		p.New = "" 
		if id != "" { //hvis id ikke er tom
			if _, idExists := lastSeen[id]; !idExists { //hvis id ikke finnes i lastSeen (map)
				p.New = id //setter p.New = id -> ny tilkobling 
				updated = true //setter updated = true
			}

			lastSeen[id] = time.Now() //setter tidspunkt for når id ble sist sett
		} //hvis id er tom, gjøres ingenting

		// Removing dead connection
		p.Lost = make([]string, 0) //lager en tom liste for tapte tilkoblinger 
		for k, v := range lastSeen { //itererer over lastSeen
			if time.Since(v) > timeout { //hvis tidspunktet for når id ble sist sett er større enn timeout = 500 * time.Millisecond = 0.5 sekunder
				updated = true 
				p.Lost = append(p.Lost, k) //legger til k (id) i p.Lost
				delete(lastSeen, k) //sletter k fra lastSeen
			}
		} 

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen)) //lager en tom liste for Peers med lengde lik antall tilkoblinger i lastSeen

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k) //legger til k (id) i p.Peers
			}

			sort.Strings(p.Peers) //sorterer p.Peers
			sort.Strings(p.Lost) //sorterer p.Lost
			peerUpdateCh <- p //sender p til peerUpdateCh
		}
	}
}