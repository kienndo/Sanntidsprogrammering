package main

import (
	"fmt"
	"sync"
	"time"
	udp "/Users/kienninhdo/Downloads/exercise4/udp"
	primary "/Users/kienninhdo/Downloads/exercise4/primary"
)


func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		primary.Primary()
	}
	wg.Wait()
}

