package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type element struct {
	mu    sync.Mutex
	value int
}

const size = 100

var data [size]element

func main() {
	fmt.Printf("cpus: %d\n", runtime.NumCPU())
	for i := 0; i < 2; i++ {
		go increaseTotals(i)
	}
	select {}
}

func increaseTotals(goroutineID int) {
	indices := make([]int, 2)
	for i := 0; ; i++ {
		seen := make(map[int]struct{}, len(indices))
		for j := 0; j < len(indices); j++ {
			for {
				n := rand.Intn(size)
				if _, ok := seen[n]; !ok {
					indices[j] = n
					seen[n] = struct{}{}
					break
				}
			}
		}
		increaseTotal(rand.Intn(i+len(indices)), indices...)
		if i%100_000 == 0 {
			fmt.Printf("%d: %v\n", goroutineID, indices)
		}
	}
}

func increaseTotal(amount int, indices ...int) {
	for _, i := range indices {
		data[i].mu.Lock()
	}

	defer func() {
		for _, i := range indices {
			data[i].mu.Unlock()
		}
	}()

	var total int
	for _, i := range indices {
		total += data[i].value
	}
	delta := amount - total
	if delta <= 0 {
		return
	}
	each := delta / len(indices)
	remainder := delta % len(indices)

	for _, i := range indices {
		data[i].value += each
		if i == len(indices)-1 {
			data[i].value += remainder
		}
	}
}
