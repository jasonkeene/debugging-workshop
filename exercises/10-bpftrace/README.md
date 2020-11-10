
## bpftrace

`bpftrace` is a tracing tool for Linux. It allows you to hook arbitrary
functions using a kernel feature called uprobes. This is similar to
`dlv trace` but with a lot less performance overhead. `bpftrace` actually has
a great deal of capability beyond this but we will focus on tracing with
uprobes specifically.

Let's measure the performance of our program with no tracing and then compare
that to Delve's tracing and `bpftrace`. First we'll need to compile the
binary to eliminate compile time from our benchmark:

```
$ cd exercises/10-bpftrace
$ go build -gcflags=all='-N -l'
```

Now let's measure a baseline with no instrumentation:

```
$ ./10-bpftrace
listening on localhost:12345
time: 459.078721ms
```

Cool, about a half of a second. Now let's measure the performance overhead of
the `ptrace` based tracepoints in Delve. What we want to measure is tracing
of the last three invocations of the `traceMe` function. Since this involves a
conditional tracepoint we can not use `dlv trace` but we can set a condition
via `dlv exec`:

```
$ dlv exec ./10-bpftrace
(dlv) trace traceMe traceMe
Tracepoint traceMe set at 0x774ff8 for main.traceMe() ./main.go:37
(dlv) cond traceMe i > 4996
(dlv)
```

By default, Delve also sets a breakpoint to print the return values of the
function:

```
(dlv) bp
Breakpoint runtime-fatal-throw at 0x43ee40 for runtime.fatalthrow() /usr/local/go/src/runtime/panic.go:1162 (0)
Breakpoint unrecovered-panic at 0x43eec0 for runtime.fatalpanic() /usr/local/go/src/runtime/panic.go:1189 (0)
        print runtime.curg._panic.arg
Tracepoint traceMe at 0x774ff8 for main.traceMe() ./main.go:37 (0)
        cond i > 4996
        args
Breakpoint 2 at 0x7750eb for main.traceMe() ./main.go:39 (0)
        args
(dlv)
```

This breakpoint will effect our benchmark results. Unfortunately, we can not
remove it but we can set its condition to `false` to effectively disable it:

```
(dlv) cond 2 false
(dlv) bp
Breakpoint runtime-fatal-throw at 0x43ee40 for runtime.fatalthrow() /usr/local/go/src/runtime/panic.go:1162 (0)
Breakpoint unrecovered-panic at 0x43eec0 for runtime.fatalpanic() /usr/local/go/src/runtime/panic.go:1189 (0)
        print runtime.curg._panic.arg
Tracepoint traceMe at 0x774ff8 for main.traceMe() ./main.go:37 (0)
        cond i > 4996
        args
Breakpoint 2 at 0x7750eb for main.traceMe() ./main.go:39 (0)
        cond false
        args
(dlv)
```

Now let's run the program and get our result:

```
(dlv) c
listening on localhost:12345
> goroutine(10): [traceMe] main.traceMe(4997, io.Writer(*net/http.response) 0xc000073828)
> goroutine(10): [traceMe] main.traceMe(4998, io.Writer(*net/http.response) 0xc000073828)
> goroutine(10): [traceMe] main.traceMe(4999, io.Writer(*net/http.response) 0xc000073828)
time: 12.088976551s
Process 30710 has exited with status 0
(dlv)
```

That is around 24 times slower! Let's see how `bpftrace` handles this.
`bpftrace` takes a `-e` argument that allows you to specify your `bpftrace`
program as a one-liner. Let's set a uprobe on the function and have it print
out a line similar to what Delve is doing:

```
$ bpftrace -e 'uprobe:./10-bpftrace:main.traceMe { if (arg2 > 4996) { printf("traceMe(%d)\n", arg2); } }'
Attaching 1 probe...
```

`bpftrace` is now monitoring any process on the system that runs the code in
the `./10-bpftrace` binary. In another terminal run the timing test:

```
$ ./10-bpftrace
listening on localhost:12345
time: 496.04543ms
```

Wow, the overhead is pretty much in the error margin! `bpftrace` has a lot of
powerful features but its support for Go is not as great as Delve's. However,
if you need low-overhead tracing it is a great option. If you want to learn
more check out the Reference Guide [here][bpftrace].

[bpftrace]: https://github.com/iovisor/bpftrace
