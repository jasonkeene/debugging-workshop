package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var someVar = "some variable"

func main() {
	fmt.Println("listening on localhost:12345")
	fmt.Printf("PID is %d\n", os.Getpid())
	log.Fatal(http.ListenAndServe("localhost:12345", http.HandlerFunc(handler)))
}

func handler(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("handler ran", someVar)
	time.Sleep(time.Second)
	rw.Write([]byte("handler response\n"))
}
