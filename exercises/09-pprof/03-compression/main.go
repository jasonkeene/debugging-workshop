package main

import (
	"compress/gzip"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	fmt.Println("listening on localhost:12345")
	go func() {
		fmt.Println(http.ListenAndServe("localhost:12345", nil))
	}()

	go doSomeCompression()
	go doSomeCompression()
	go doSomeCompression()
	doSomeCompression()
}

func doSomeCompression() {
	for {
		w := gzip.NewWriter(ioutil.Discard)
		_, err := io.Copy(w, io.LimitReader(rand.Reader, 4194304))
		if err != nil {
			fmt.Println(err)
		}
		err = w.Close()
		if err != nil {
			fmt.Println(err)
		}
	}
}
