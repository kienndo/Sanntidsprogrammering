package backup

import(
	"net"
	"fmt"
	"time"
	"math/rand"
)

func StartPrimary() {

	addr, err := net.ResolveUDPAddr("udp", "localhost:15657")
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Initialize the random seed
    rand.Seed(time.Now().UnixNano())

    // Generate a random duration between 0 and 10 seconds
    randomDuration := time.Millisecond * time.Duration(rand.Intn(10000))

	buffer := make([]byte, 1024)

    // Set the read deadline from the current time plus the random duration
    err = conn.SetReadDeadline(time.Now().Add(randomDuration))
	_,_, err = conn.ReadFromUDP(buffer)
    if err != nil {
	    fmt.Println("Failed to set read deadline:", err)
	}

	if err != nil {
		fmt.Println("No message recieved, becoming primary")
		conn.Close()
		return
	}



}