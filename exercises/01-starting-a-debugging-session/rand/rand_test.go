package rand

import (
	"runtime"
	"testing"
)

func TestRandomNumber(t *testing.T) {
	runtime.Breakpoint()

	rn := RandomNumber()

	if rn != 4 {
		t.Fatal("this is clearly not a random number")
	}
}
