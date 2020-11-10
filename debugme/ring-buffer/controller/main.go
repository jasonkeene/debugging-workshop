package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-delve/delve/service/api"
	"github.com/go-delve/delve/service/rpc2"
)

func main() {
	// addr for the JSON RPC API server to bind to
	const addr = "localhost:12345"

	// notify when signal is sent to exit
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	// run delve in headless mode
	cmd := exec.Command(
		"dlv",
		"debug",
		"--headless",
		"--listen="+addr,
		"--api-version=2",
		"--accept-multiclient",
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = ".."
	cmd.Start()
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// wait for delve to start accepting connections
	waitToBeUp(addr)

	// create a client and continue the process
	client := rpc2.NewClient(addr)
	client.Continue()
	defer func() {
		client.Halt()
		client.Detach(false)
	}()

	// find program counter
	locations, err := client.FindLocation(
		api.EvalScope{
			GoroutineID: -1,
		},
		"main.(*ringBuffer).read",
		true,
	)
	if err != nil {
		panic(err)
	}
	pc := locations[0].PC

	// create breakpoint
	bp := &api.Breakpoint{
		Name: "readIndices",
		Addr: pc,
		Variables: []string{
			"b.writeIndex",
			"b.readIndex",
		},
	}

	for {
		// check to see if we should exit
		select {
		case <-done:
			return
		default:
		}

		client.Halt()
		client.CreateBreakpoint(bp)
		state := <-client.Continue()
		client.ClearBreakpointByName(bp.Name)
		client.Continue()

		writeIndex, err := strconv.Atoi(state.CurrentThread.BreakpointInfo.Variables[0].Value)
		if err != nil {
			panic(err)
		}
		readIndex, err := strconv.Atoi(state.CurrentThread.BreakpointInfo.Variables[1].Value)
		if err != nil {
			panic(err)
		}

		fmt.Printf("\nw: %d r: %d delta: %d\n", writeIndex, readIndex, writeIndex-readIndex)

		time.Sleep(time.Second)
	}
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
