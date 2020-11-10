package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
)

func main() {
	fmt.Println("listening on localhost:12345")
	go func() {
		panic(http.ListenAndServe("localhost:12345", http.HandlerFunc(handler)))
	}()

	waitToBeUp("localhost:12345")

	start := time.Now()
	for i := 0; i < 5000; i++ {
		resp, err := http.Get(fmt.Sprintf("http://localhost:12345/%d", i))
		if err != nil {
			panic(err)
		}
		ioutil.ReadAll(resp.Body)
	}
	fmt.Printf("time: %s\n", time.Now().Sub(start))
}

func handler(rw http.ResponseWriter, req *http.Request) {
	i, _ := strconv.Atoi(req.URL.Path[1:])
	traceMe(i, rw)
}

func traceMe(i int, w io.Writer) {
	fmt.Fprintf(w, "handled response for %d\n", i)
}

func waitToBeUp(addr string) {
	done := time.After(5 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			conn.Close()
			return
		}
		select {
		case <-done:
			panic(err)
		default:
		}
	}
}
