package backup

import (
	"fmt"
	"time"
)

var (
	PrimaryAlive   bool
	LastNumber     int
	CountingNumber int
)

func Increment(lastNumber int) {
	for {
		for i := lastNumber + 1; i < 6; i++ {
			CountingNumber++
			fmt.Printf("%d\n", CountingNumber)
			time.Sleep(time.Second)

			if CountingNumber == 4 {
				CountingNumber = 0
			}
		}

	}
}

func Backup() {
	PrimaryAlive, LastNumber := UDPrecieve()

	if PrimaryAlive != true {
		go Increment(LastNumber)
	}
}
