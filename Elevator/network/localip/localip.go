package localip

import (
	"net"
	"strings"
)

var localIP string 

func LocalIP() (string, error) {
	if localIP == "" { //hvis localIP er tom
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53}) //lager en TCP-tilkobling til 
		if err != nil { 
			return "", err
		}
		defer conn.Close() //lukker tilkoblingen
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0] //splitter LocalAddr().String() på ":" og lagrer første del i localIP når vi har hentet IP-adressen
	}
	return localIP, nil
}
