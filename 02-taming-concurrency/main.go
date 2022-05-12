package main

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	ch := make(chan int)

	go func() {
		defer close(ch)

		for {
			v := rand.Intn(100)

			if v == 42 {
				return
			}

			ch <- v
		}
	}()

	sum := 0
	for v := range ch {
		sum += v
	}

	fmt.Printf("sum is %d\n", sum)
}
