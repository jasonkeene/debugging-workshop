
## Debugging From Core Dumps

A core dump is a serialized form of the current execution state of the
process. The name "core dump" originates from "core memory" which was a type
of magnetic memory used in early computers before silicon memory was created.

These dumps include the of process's memory, registers, flags, and some
operating system specific state. We can load these files into Delve and
inspect them.

⚠️ Since these dumps contain all the process's memory it is critical that they
are handled securely. If there is confidential information in memory at the
time the core dump is taken this will be part of the core dump.

There are two situations where you might want to debug from a core file.

### Post-mortem Debugging

When your process crashes you can tell Go to write out a core dump in
addition to the regular stack trace. This can be useful to debug what caused
the crash. This is disabled by default on many systems as core dumps can take
up a lot of disk space and can reduce system performance when crashes occur.

You will need to ensure the limit for core dump file size is set
appropriately:

```
$ ulimit -c
0
```

If the value is set to `0` then core dumps are disabled and will need to be
enabled. You can enable core dumps for a given shell by running:

```
$ ulimit -c unlimited
$ ulimit -c
unlimited
```

However, this setting is specific to this shell. To set the limit for all
users you will need to modify `/etc/security/limits.conf`:

```
$ sudo -i
# echo '* soft core unlimited' >> /etc/security/limits.conf
```

Now when you restart your shell the limit will be set to `unlimited`. This is
suitable for a development system, however, the size of core dumps should be
constrained on systems that do real work.

Our system will now record core dumps, however, we need to tell the Go
runtime to produce a core dump on crashes. To do this we run our process with
the `GOTRACEBACK` env var set to `crash`. Let's try this:

```
$ cd exercises/06-debugging-from-core-dumps
$ go build -o server main.go
$ GOTRACEBACK=crash ./server
listening on :12345
PID is 12684
```

Now we can curl the server and cause it to crash:

```
$ curl localhost:12345/crash
curl: (52) Empty reply from server
```

You should see the server output the following message:

```
panic: crashing!

goroutine 8 [running]:
... stacktrace ...

Aborted (core dumped)
```

You should now see a file called `core` in the current working directory.
Let's debug it!

```
$ dlv core server core
(dlv) goroutines
  Goroutine 1 - User: /usr/local/go/src/runtime/netpoll.go:220 internal/poll.runtime_pollWait (0x465c35)
  Goroutine 2 - User: /usr/local/go/src/runtime/proc.go:307 runtime.gopark (0x439c65)
  Goroutine 3 - User: /usr/local/go/src/runtime/proc.go:307 runtime.gopark (0x439c65)
  Goroutine 4 - User: /usr/local/go/src/runtime/proc.go:307 runtime.gopark (0x439c65)
  Goroutine 5 - User: /usr/local/go/src/runtime/proc.go:307 runtime.gopark (0x439c65)
  Goroutine 6 - User: /usr/local/go/src/runtime/time.go:188 time.Sleep (0x467e7f)
  Goroutine 7 - User: /usr/local/go/src/runtime/netpoll.go:220 internal/poll.runtime_pollWait (0x465c35)
* Goroutine 8 - User: ./main.go:21 main.handler.func1 (0x634739) (thread 12684)
[8 goroutines]
(dlv) bt
 0  0x000000000046c6a1 in runtime.raise
    at /usr/local/go/src/runtime/sys_linux_amd64.s:165
 1  0x000000000044c7fd in runtime.dieFromSignal
    at /usr/local/go/src/runtime/signal_unix.go:754
 2  0x000000000044cd11 in runtime.sigfwdgo
    at /usr/local/go/src/runtime/signal_unix.go:968
 3  0x000000000044b4b4 in runtime.sigtrampgo
    at /usr/local/go/src/runtime/signal_unix.go:409
 4  0x000000000046ca43 in runtime.sigtramp
    at /usr/local/go/src/runtime/sys_linux_amd64.s:409
 5  0x00007fc58e6e78a0 in ???
    at ?:-1
 6  0x000000000043732b in runtime.crash
    at /usr/local/go/src/runtime/signal_unix.go:846
 7  0x000000000043732b in runtime.fatalpanic
    at /usr/local/go/src/runtime/panic.go:1216
 8  0x0000000000436c90 in runtime.gopanic
    at /usr/local/go/src/runtime/panic.go:1064
 9  0x0000000000634739 in main.handler.func1
    at ./main.go:21
10  0x000000000046ae81 in runtime.goexit
    at /usr/local/go/src/runtime/asm_amd64.s:1374
(dlv) frame 9
> runtime.raise() /usr/local/go/src/runtime/sys_linux_amd64.s:165 (PC: 0x46c6a1)
Warning: debugging optimized function
Frame 9: ./main.go:21 (PC: 634739)
    16:
    17: func handler(rw http.ResponseWriter, req *http.Request) {
    18:         fmt.Println("handler ran")
    19:         if req.URL.Path == "/crash" {
    20:                 go func() {
=>  21:                         panic("crashing!")
    22:                 }()
    23:         }
    24:         time.Sleep(time.Second)
    25:         rw.Write([]byte("handler response\n"))
    26: }
```

So we can see the crash happened in our handler.

Alternatively, instead of waiting around for a crash to happen, you can induce
one by sending the `SIGQUIT` signal. If we run the server again:

```
$ GOTRACEBACK=crash ./server
listening on :12345
PID is 12754
```

We can send the signal by pressing `Ctrl+\`:

```
^\SIGQUIT: quit
PC=0x46cf00 m=0 sigcode=128

goroutine 0 [idle]:
... stacktrace ...

Aborted (core dumped)
```

### Snapshot Debugging

In addition, to using core dumps for post-mortem debugging you can obtain a
core dump from a process without it crashing. These are known as snapshots.
These can be useful when you don't otherwise want to attach a debugger
directly to the running process. Instead, you can take a snapshot and debug the
issue on your own time without halting the process.

Let's run the server again:

```
$ ./server
listening on :12345
PID is 12765
```

We can then create a core dump by providing `gcore` the PID of the process:

```
$ gcore 12765
[New LWP 12766]
[New LWP 12767]
[New LWP 12768]
[New LWP 12769]
[Thread debugging using libthread_db enabled]
Using host libthread_db library "/lib/x86_64-linux-gnu/libthread_db.so.1".
runtime.epollwait () at /usr/local/go/src/runtime/sys_linux_amd64.s:725
725             MOVL    AX, ret+24(FP)
Saved corefile core.12765
```

`gcore` must suspend the process while it is taking the core dump. As a
result, the process might be suspended for 5 or so seconds depending on the
core dump size. However, after you pay this one-time penalty you can now
debug the core dump without impeding the process.

For more information about core dumps see the [manpage][manpage].

[manpage]: https://man7.org/linux/man-pages/man5/core.5.html

### DEBUGME: Core Dump

In the directory `debugme/resolve-hostname` there is a core dump file. See if
you can find out what happened from the core dump alone.
