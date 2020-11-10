package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	fmt.Println("listening on localhost:12345")
	go func() {
		fmt.Println(http.ListenAndServe("localhost:12345", nil))
	}()

	for {
		time.Sleep(10 * time.Millisecond)
		select {
		case resp := <-getResource():
			if resp.StatusCode < 400 {
				print(".")
			} else {
				print("x")
			}
		case <-time.After(75 * time.Millisecond):
			print("X")
		}
	}
}

func getResource() <-chan *http.Response {
	ch := make(chan *http.Response)
	go func() {
		resp, _ := http.Get("https://www.google.com/")
		ch <- resp
	}()
	return ch
}
