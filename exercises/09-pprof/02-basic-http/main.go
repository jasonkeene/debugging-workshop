package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	fmt.Println("listening on localhost:12345")
	fmt.Println(http.ListenAndServe("localhost:12345", nil))
}
