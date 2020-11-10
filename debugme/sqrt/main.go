package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("argument requred")
		os.Exit(1)
	}

	in, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		fmt.Println("argument was not a number")
		os.Exit(1)
	}

	fmt.Println(sqrt(in).String())
}

type result struct {
	sqrt float64
	i    bool
}

func (r *result) String() string {
	var i string
	if r.i {
		i = "i"
	}
	return fmt.Sprintf("%g%s", r.sqrt, i)
}

var memoizedSqrt = make(map[float64]*result)

func sqrt(in float64) *result {
	r, ok := memoizedSqrt[in]
	if ok {
		return r
	}

	if in >= 0 {
		r = &result{
			sqrt: math.Sqrt(in),
		}
		memoizedSqrt[in] = r
	} else {
		r := &result{
			sqrt: math.Sqrt(math.Abs(in)),
			i:    true,
		}
		memoizedSqrt[in] = r
	}

	return r
}
