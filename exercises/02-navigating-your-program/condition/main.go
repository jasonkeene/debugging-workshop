package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	fmt.Println("listening on localhost:12345")
	fmt.Printf("PID is %d\n", os.Getpid())
	log.Fatal(http.ListenAndServe("localhost:12345", http.HandlerFunc(handler)))
}

func handler(rw http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal("request handled")
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(append(data, '\n'))
}
