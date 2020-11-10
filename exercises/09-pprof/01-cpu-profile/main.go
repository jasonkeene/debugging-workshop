package main

import (
	"crypto/sha256"
	"log"
	"os"
	"runtime/pprof"
)

func main() {
	f, err := os.Create("./cpu.prof.pb.gz")
	if err != nil {
		log.Fatal(err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	data := []byte("original data")
	for i := 0; i < 1_000; i++ {
		sum := sha256.Sum256(data)
		data = sum[:]
	}
}
