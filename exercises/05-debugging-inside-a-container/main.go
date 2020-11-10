package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("listening on :12345")
	log.Fatal(http.ListenAndServe(":12345", http.HandlerFunc(handler)))
}

func handler(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("handler ran")
	rw.Write([]byte("handler ran\n"))
}
