package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

func main() {
	addr, err := startServer()
	if err != nil {
		log.Fatal(err)
	}

	sendRequest(addr)
}

func startServer() (string, error) {
	l, err := net.Listen("tcp", "localhost:")
	if err != nil {
		return "", err
	}

	go func() {
		err := http.Serve(l, http.HandlerFunc(handler))
		if err != nil {
			log.Print(err)
		}
	}()

	return l.Addr().String(), nil
}

func handler(rw http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal("request handled")
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(append(data, '\n'))
}

func sendRequest(addr string) {
	resp, err := http.Get(fmt.Sprintf("http://%s", addr))
	if err != nil {
		log.Print(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}
	fmt.Printf("resp: %d %s\n", resp.StatusCode, data)
}
