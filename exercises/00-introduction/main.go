package main

import "runtime"

func main() {
	println("this runs before the static breakpoint")

	// This is an example of adding a static breakpoint in your program.
	runtime.Breakpoint()

	// The debugger will stop execution before this line and so you will not see
	// this output immediately.
	println("this runs after the static breakpoint")

	println("a bunch")
	println("of other")
	println("lines")

	// BREAK HERE
	println("add a breakpoint on this line")
}
