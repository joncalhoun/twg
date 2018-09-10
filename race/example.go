package main

import (
	"fmt"
	"sync"
	"time"
)

var balance = 100

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	go spend(30, &wg)
	go spend(40, &wg)
	wg.Wait()
	fmt.Println(balance)
}

func spend(amount int, wg *sync.WaitGroup) {
	b := balance
	time.Sleep(time.Second)
	b -= amount
	balance = b
	wg.Done()
}
