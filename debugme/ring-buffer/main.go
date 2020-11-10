package main

import (
	"fmt"
	"sync"
)

type ringBuffer struct {
	readIndex  uint64
	writeIndex uint64

	mu     sync.RWMutex
	buffer []int
}

func (b *ringBuffer) read() int {
	// block until there is data to read
	for {
		b.mu.RLock()
		if b.readIndex < b.writeIndex {
			defer b.mu.RUnlock()
			break
		}
		b.mu.RUnlock()
	}

	v := b.buffer[b.readIndex%uint64(len(b.buffer))]
	b.readIndex++
	return v
}

func (b *ringBuffer) write(v int) {
	// block until there is a spot available to do the write
	for {
		b.mu.Lock()
		if b.writeIndex <= b.readIndex+uint64(len(b.buffer)) {
			defer b.mu.Unlock()
			break
		}
		b.mu.Unlock()
	}

	b.buffer[b.writeIndex%uint64(len(b.buffer))] = v
	b.writeIndex++
}

func main() {
	const size = 20000

	buf := ringBuffer{
		buffer: make([]int, size),
	}

	// goroutine to do writes, in-order
	go func() {
		for i := 0; ; i++ {
			buf.write(i)
		}
	}()

	// do reads, hopefully they are also in-order
	prev := -1 // previous read
	for {
		v := buf.read()

		// if current read is not +1 from the previous read warn the user
		if v != prev+1 {
			fmt.Print("WARNING: data is not in order or data is missing\n")
		}

		// progress indicator
		if v%size == 0 {
			fmt.Print(".")
		}

		prev = v
	}
}
