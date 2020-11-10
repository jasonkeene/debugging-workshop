package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jasonkeene/debugging-workshop/exercises/01-starting-a-debugging-session/rand"
)

func main() {
	fmt.Println("listening on localhost:12345")
	fmt.Printf("PID is %d\n", os.Getpid())
	log.Fatal(http.ListenAndServe("localhost:12345", http.HandlerFunc(handler)))
}

func handler(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("providing a random number")
	fmt.Fprintf(rw, "%d\n", rand.RandomNumber())
}
