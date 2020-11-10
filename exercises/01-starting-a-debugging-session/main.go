package main

import (
	"runtime"

	"github.com/jasonkeene/debugging-workshop/exercises/01-starting-a-debugging-session/rand"
)

func main() {
	runtime.Breakpoint()

	println(rand.RandomNumber())
}
