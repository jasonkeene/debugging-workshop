package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	ips, err := resolve(hostname)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", ips)
}

type codeError struct {
	error
	code int
}

func resolve(hostname string) ([]string, *codeError) {
	ips, err := net.LookupHost(hostname)
	if err != nil {
		return nil, &codeError{
			error: err,
			code:  500,
		}
	}
	return ips, nil
}
