
## pprof

`pprof` allows you to collect and analyze profile data from your Go programs.
These profiles consist of a series of sampled values that are associated with
stack traces in your program. These profiles are stored as a gziped protobuf
file. pprof also supports a text format that is human readable.

The pprof format does not enforce any semantics on the profile data, as such,
we can use this format to generate different sorts of profiles. These are the
profiles available by default:

- **cpu:** This profile is different than the rest. It samples the stack
  traces of all active threads every 10ms. It is also known as just "profile"
  as it is the default profile.
- **goroutine:** This profile takes a point-in-time snapshot of the stack
  traces of all current goroutines.
- **heap:** This profile shows memory allocations for both in-use memory
  (allocs that were not freed) and previous allocations (allocs that were
  freed). By default this profile displays in-use space.
- **allocs:** This profile is the same as the heap profile but defaults to
  showing previous allocations.
- **threadcreate:** - This profile shows stack traces that led to the
  creation of new OS threads.
- **block:** This profile shows stack traces that led to blocking on
  synchronization primitives.
- **mutex:** This profile shows stack traces of holders of contended mutexes.

The default profiles cover a lot of ground out of the box, however, you can
also add your own custom profiles. For instance, if you wanted to profile open
files, sockets, or some other resource. Really, anything that you want to
measure with a stack trace pprof can help you with.

The advantage of capturing profiles is that they can allow for offline
debugging of production processes similar to taking a core dump but with much
less performance overhead and fewer concerns around security.

### Capturing Profiles

Since there is typically some performance overhead with taking profiles you
want to only capture profiles when needed. To capture a CPU profile we just
need to call a few functions to start and stop the profile:

```go
pprof.StartCPUProfile(someWriter)
// do some work
pprof.StopCPUProfile()
```

This will capture all on CPU work being done. Let's experiment with doing this:

```
$ cd exercises/09-pprof/01-cpu-profile
$ cat main.go
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
```

This program takes a CPU profile and writes it out to `cpu.prof.pb.gz`.
During the CPU profile we do a bunch of hashing as an example. Let's run this:

```
$ go run .
$ ls
cpu.prof.pb.gz  main.go
```

We can see what is in the profile by running:

```
$ go tool pprof -top cpu.prof.pb.gz
Type: cpu
Time: Nov 7, 2020 at 6:52pm (MST)
Duration: 205.90ms, Total samples = 0
Showing nodes accounting for 0, 0% of 0 total
      flat  flat%   sum%        cum   cum%
```

Wait, that is not right, we should see the CPU being used to do some hashing.
The reason we don't see anything here is that our workload did not run long
enough to be within a sample. The CPU profiler takes samples once every 10ms
and if our workload is not running when it is ready to sample it won't be
recorded. This demonstrates the downside of sampling, however, it is a
trade-off between performance impact and fidelity.

We can increase the duration of the workload by doing 1,000,000 iterations vs
1,000. We can now see some results:

```
$ go run .
$ go tool pprof -top cpu.prof.pb.gz
Type: cpu
Time: Nov 7, 2020 at 6:57pm (MST)
Duration: 401.07ms, Total samples = 220ms (54.85%)
Showing nodes accounting for 220ms, 100% of 220ms total
      flat  flat%   sum%        cum   cum%
     110ms 50.00% 50.00%      110ms 50.00%  crypto/sha256.block
      20ms  9.09% 59.09%       20ms  9.09%  runtime.(*mspan).init (inline)
      20ms  9.09% 68.18%       20ms  9.09%  runtime.kevent
      20ms  9.09% 77.27%       20ms  9.09%  runtime.madvise
      10ms  4.55% 81.82%       10ms  4.55%  crypto/sha256.(*digest).Reset
      10ms  4.55% 86.36%      120ms 54.55%  crypto/sha256.(*digest).Write
      10ms  4.55% 90.91%      140ms 63.64%  crypto/sha256.Sum256
      10ms  4.55% 95.45%       30ms 13.64%  runtime.mallocgc
      10ms  4.55%   100%       10ms  4.55%  runtime.pthread_kill
```

Other profiles besides the CPU profile can be captured in a similar, but
different way:

```
pprof.Lookup("goroutine").WriteTo(someWriter, 0)
pprof.Lookup("threadcreate").WriteTo(someWriter, 0)
pprof.Lookup("heap").WriteTo(someWriter, 0)
pprof.Lookup("allocs").WriteTo(someWriter, 0)
pprof.Lookup("block").WriteTo(someWriter, 0)
pprof.Lookup("mutex").WriteTo(someWriter, 0)
```

You can also easily capture CPU and memory profiles of your tests and
benchmarks by running them with a few flags:

```
go test -cpuprofile cpu.prof -memprofile mem.prof .
go test -cpuprofile cpu.prof -memprofile mem.prof -bench .
```

### `pprof` Tool

As we just saw, go packages the `pprof` tool as `go tool pprof`. This is a
vendored version of `github.com/google/pprof` that is tested to work with
your version of Go. This tool can be used to analyze profiles. You can
generate graphs in a number of formats such as PDF:

```
$ go tool pprof -pdf cpu.prof.pb.gz
Generating report in profile001.pdf
```

You can also show your source code annotated with how much time is being
spent on each line:

```
$ go tool pprof -list main.main cpu.prof.pb.gz
Total: 210ms
ROUTINE ======================== main.main in /Users/jasonkeene/src/github.com/jasonkeene/debugging-workshop/exercises/09-pprof/01-cpu-profile/main.go
      20ms      200ms (flat, cum) 95.24% of Total
         .          .     18:   }
         .          .     19:   defer pprof.StopCPUProfile()
         .          .     20:
         .          .     21:   data := []byte("original data")
         .          .     22:   for i := 0; i < 1_000; i++ {
      20ms      200ms     23:           sum := sha256.Sum256(data)
         .          .     24:           data = sum[:]
         .          .     25:   }
         .          .     26:}
```

You can even disassemble a function and view the instructions that are taking
the most time. You need to provide pprof the binary to do this:

```
$ go build -o tmp
$ go tool pprof -disasm Write tmp cpu.prof.pb.gz
...
         .      130ms    108d128: CALL crypto/sha256.block(SB)            ;crypto/sha256.(*digest).Write sha256.go:191
...
```

We can see the call to `crypto/sha256.block` is taking up a lot of time.

If you do not provide `pprof` a sub-command it will land you in a an
interactive mode.

```
$ go tool pprof cpu.prof.pb.gz
Type: cpu
Time: Nov 7, 2020 at 7:04pm (MST)
Duration: 400.93ms, Total samples = 210ms (52.38%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top10
Showing nodes accounting for 210ms, 100% of 210ms total
Showing top 10 nodes out of 19
      flat  flat%   sum%        cum   cum%
     130ms 61.90% 61.90%      130ms 61.90%  crypto/sha256.block
      20ms  9.52% 71.43%      160ms 76.19%  crypto/sha256.(*digest).checkSum
      20ms  9.52% 80.95%      200ms 95.24%  main.main
      10ms  4.76% 85.71%      150ms 71.43%  crypto/sha256.(*digest).Write
      10ms  4.76% 90.48%       10ms  4.76%  runtime.(*mcentral).grow
      10ms  4.76% 95.24%       10ms  4.76%  runtime.madvise
      10ms  4.76%   100%       10ms  4.76%  runtime.memmove
         0     0%   100%      170ms 80.95%  crypto/sha256.Sum256
         0     0%   100%       10ms  4.76%  runtime.(*mcache).nextFree
         0     0%   100%       10ms  4.76%  runtime.(*mcache).refill
(pprof)
```

### Web Interface

Go provides an easy way to expose an HTTP interface to pprof via the
`net/http/pprof` package. We just need to import the package:

```go
import _ "net/http/pprof"
```

This will wire up some handlers to the `DefaultServeMux` on init. That
happens [here][pprof-init].

And at some point we need to use the `DefaultServeMux` to handle requests.
That can be done by calling `http.ListenAndServe` with no handler:

```go
http.ListenAndServe("localhost:6060", nil)
```

This is all that is needed in order to start capturing profiles. Let's run a
basic program with pprof enabled:

```
$ cd exercises/09-pprof/02-basic-http
$ go run .
listening on localhost:12345
```

We can then capture a profile by simply hitting the `profile` endpoint:

```
$ curl -s localhost:12345/debug/pprof/profile?seconds=2 | od -a
0000000   us  8b  bs nul nul nul nul nul eot    nul   O nul         H
0000020   90                 89    syn  nl eot  bs soh dle stx  nl
0000040  eot  bs etx dle eot   P     8b       bel   Z eot  bs etx dle
0000060  eot   `  80       eot sub stx  bs soh   2 nul   2 bel   s   a
0000100    m   p   l   e   s   2 enq   c   o   u   n   t   2 etx   c   p
0000120    u   2  vt   n   a   n   o   s   e   c   o   n   d   s soh nul
0000140  nul             can      O nul nul nul
0000153
```

This generates a protobuf message that contains the profile data. There is
not much to see here since our program isn't doing much but it demonstrates
how to take a CPU profile. We specify the `seconds` argument otherwise it
will default to a 30 second profile.

You can visit `http://localhost:12345/debug/pprof/` in your browser. This
gives you links and descriptions about what sort of profiles you can take.

In additional to built-in pprof profiles there are a few other endpoints you
can access via the HTTP interface. For instance `cmdline` will give you the
command line args:

```
$ curl http://localhost:12345/debug/pprof/cmdline
/var/folders/wc/smsj4m2935jdh3zn7tkjdqb00000gn/T/go-build976420283/b001/exe/02-basic-http
```

There is a `symbol` endpoint that allows you to easily lookup symbols from
raw addresses:

```
$ curl http://localhost:12345/debug/pprof/symbol?0x13723c0
num_symbols: 1
0x13723c0 main.main
```

Finally, there is a `trace` endpoint that will capture a one second trace. You
can increase the time the trace runs for with the `seconds` query arg. These
traces can be analyzed by `go tool trace` which launches it's own web UI
specific to inspecting traces.

Let's use `go tool pprof` to capture a CPU profile and display a web UI to
inspect it. For this let's run a CPU bound workload:

```
$ cd exercises/09-pprof/03-compression
$ go run .
```

Then in another terminal:

```
$ go tool pprof -http=: -seconds=2 http://localhost:12345/debug/pprof/profile
Fetching profile over HTTP from http://localhost:12345/debug/pprof/profile?seconds=2
Please wait... (2s)
Saved profile in /Users/jasonkeene/pprof/pprof.samples.cpu.005.pb.gz
Serving web UI on http://localhost:54448
```

You should now see a graph showing the stack frames that were active when the
samples were taken for the profile. You can see we are spending most of our
time doing compression and reading random data from the kernel. In addition
to the graph, there is a top and flame graph view I find useful that displays
the same information in a different format.

### DEBUGME: Goroutine and Heap Profiles

There are two programs under `debugme/goroutine-leak-2` and
`debugme/memory-leak`. Try your hand at debugging them. Use the goroutine
profile for the goroutine leak and the heap profile for the memory leak.

[pprof-init]: https://github.com/golang/go/blob/e1b305af028544e00a22c905e68049c98c10a1cc/src/net/http/pprof/pprof.go#L80-L86
