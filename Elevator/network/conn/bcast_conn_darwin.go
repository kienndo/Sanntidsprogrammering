// +build darwin

package conn

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

func DialBroadcastUDP(port int) net.PacketConn {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP) //lager socket
	if err != nil { 
		fmt.Println("Error: Socket:", err) 
	}
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1) //socket options: tillater at flere kan lytte på samme port
	if err != nil { 
		fmt.Println("Error: SetSockOpt REUSEADDR:", err) 
	}
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1) //socket options: tillater sending av broadcast-meldinger
	if err != nil { 
		fmt.Println("Error: SetSockOpt BROADCAST:", err) 
	}
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1) //tillater flere prosesser å bruke samme port
	if err != nil { 
		fmt.Println("Error: SetSockOpt REUSEPORT:", err) 
	}
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port}) //binder socket til port {Port: port}
	if err != nil { 
		fmt.Println("Error: Bind:", err) 
	}
	f := os.NewFile(uintptr(s), "") //s = socket, lager en ny fil fra socketen
	conn, err := net.FilePacketConn(f) //lager en net.PacketConn av socket-filen -> brukes for å sende og motta data
	if err != nil { 
		fmt.Println("Error: FilePacketConn:", err) 
	}
	f.Close()

	return conn //returnerer net.PacketConn: nå kan vi sende UDP pakker og broadcaste til flere Receivers:)
}
