package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/tink"
)

var a tink.AEAD

func init() {
	kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	if err != nil {
		log.Fatal(err)
	}

	a, err = aead.New(kh)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Fatal(http.ListenAndServe("localhost:54321", http.HandlerFunc(handler)))
}

// handler encrypts the request body and returns it as a response
func handler(rw http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	ct, err := a.Encrypt(data, nil)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Write(ct)
}
