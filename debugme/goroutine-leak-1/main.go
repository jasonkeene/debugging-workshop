package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

// This program runs an HTTP server and every second sends an HTTP request to
// the server. Unfortunately, when I wrote this I introduced a goroutine leak!
// Can you use Delve to detect the goroutine leak? Can you find the cause of the
// leak? What is the fix?

func main() {
	addr, done, err := startServer()
	if err != nil {
		log.Fatal(err)
	}

	sendRequests(addr, done)
}

func startServer() (string, <-chan struct{}, error) {
	l, err := net.Listen("tcp", "localhost:")
	if err != nil {
		return "", nil, err
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		err := http.Serve(l, http.HandlerFunc(handler))
		if err != nil {
			log.Print(err)
		}
	}()

	return l.Addr().String(), done, nil
}

func handler(rw http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal("request handled")
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(append(data, '\n'))
}

func sendRequests(addr string, done <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			_, err := http.Get(fmt.Sprintf("http://%s", addr))
			if err != nil {
				log.Print(err)
			}
		case <-done:
			return
		}
	}
}
