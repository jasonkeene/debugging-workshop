package main

import (
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
)

const prefixLen = 6

var letters = []rune("abc")

func main() {
	fmt.Println("listening on localhost:12345")
	go func() {
		fmt.Println(http.ListenAndServe("localhost:12345", nil))
	}()

	// There is a maximum of len(letters)^prefixLen prefix keys (729).
	count := make(map[string]int, 729)

	for i := 0; ; i++ {
		count[prefix(randomString())]++
		if i%100 == 0 {
			fmt.Println(len(count))
		}
	}
}

func prefix(s string) string {
	return s[:prefixLen]
}

func randomString() string {
	b := make([]rune, 1<<20)
	for i := 0; i < prefixLen; i++ {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
