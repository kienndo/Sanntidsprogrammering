package primary

import (
	"fmt"
	udp "exercise4/udp"
)

var (
	CountingNumber = 0
)

func Increment() {
	for {
		for i := 0; i < 6; i++ {
			CountingNumber++
			fmt.Printf("%d\n", CountingNumber)
			time.Sleep(time.Second)

			if CountingNumber == 4 {
				CountingNumber = 0
			}
		}

	}
}

func Primary(){

	go Increment()
	go UDPsend(CountingNumber)

}