package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Semaphore struct{}
type URL int
type Result struct {
	Value string
	//Leaving space for other things
}

func (r Result) String() string {
	return r.Value
}

func random(max int) int {
	return (rand.Int() % max)
}

func producer(results chan Result, work URL) {
	doze := time.Duration(random(500)) * time.Millisecond
	time.Sleep(doze)
	r := Result{
		Value: fmt.Sprintf("ProcessedURL %d, slept %s", work, doze),
	}
	results <- r
	return
}

func satisfied(i int) bool {
	//We're launching
	if i > 9 {
		return true
	}

	return false
}

func consumer(results chan Result, done chan Semaphore) {
	i := 0
	for !satisfied(i) {
		i++
		select {
		case res, ok := <-results:
			if !ok {
				fmt.Printf("I is %d but the channel is closed.", i)
				done <- Semaphore{}
				return
			}
			fmt.Printf("%s is overall #%d\n", res, i)

			if !satisfied(i) {
				//If it's not happy after this result, the consumer
				// instructs a producer to start on something new
				job := URL(random(300))
				fmt.Printf("Consumer is not satisfied after job #%d. Launching %d\n", i, job)
				go producer(results, job)
			} else {
				fmt.Printf("Consumer is satisfied after job #%d. Unlocking.\n", i)
				done <- Semaphore{}
			}
		}
	}
}

func main() {
	simultaneous := 5

	//Buffered channel of results
	results := make(chan Result, simultaneous)
	blocker := make(chan Semaphore)

	//Pull down "URLs" simultaneously to kick things off
	for i := 0; i < simultaneous; i++ {
		x := random(300)
		fmt.Printf("Launching %d\n", x)
		go producer(results, URL(x))
	}
	go consumer(results, blocker)

	//Wait until the consumer is satisfied
	<-blocker

	//Drain the channel
	<-results
	close(results)

	fmt.Printf("Done\n")
}
