
## Starting a Debugging Session

There are a number of ways to start a debugging session with Delve.

### `debug`

If you would like to debug a `main` package you can use `dlv debug`. You saw
how this works in the previous exercise. You can optionally give `dlv debug`
an argument which is the package you'd like to debug. For instance, if you
want to use Delve to debug Delve itself you can do:

```
$ dlv debug github.com/go-delve/delve/cmd/dlv
```

This may tear open a hole in the fabric of spacetime though, so proceed with
caution.

### `test`

`dlv test` allows you to debug your tests. Let's debug the tests in the `rand`
package!

```
$ cd exercises/01-starting-a-debugging-session
$ dlv test ./rand
Type 'help' for list of commands.
(dlv) c
> github.com/jasonkeene/debugging-workshop/exercises/01-starting-a-debugging-session/rand.TestRandomNumber() ./rand/rand_test.go:11 (PC: 0x1166346)
     6: )
     7:
     8: func TestRandomNumber(t *testing.T) {
     9:         runtime.Breakpoint()
    10:
=>  11:         rn := RandomNumber()
    12:
    13:         if rn != 4 {
    14:                 t.Fatal("this is clearly not a random number")
    15:         }
    16: }
(dlv) c
PASS
Process 82900 has exited with status 0
```

### `exec`

You might already have your program compiled and you want Delve to run it and
attach to it. In that case you can use `delve exec`. Let's try doing that:

```
$ go build
$ dlv exec ./01-starting-a-debugging-session
Type 'help' for list of commands.
(dlv) c
> main.main() ./main.go:12 (PC: 0x105c625)
Warning: debugging optimized function
     7: )
     8:
     9: func main() {
    10:         runtime.Breakpoint()
    11:
=>  12:         println(rand.RandomNumber())
    13: }
(dlv) c
4
Process 84664 has exited with status 0
```

Using `delve exec` has some downsides, however. Since Delve is not compiling
the program, by default Go will inline certain functions and perform other
optimizations that make for a less than ideal debugging experience. For
instance our `rand.RandNumber()` function, while guaranteed to be random,
will be inlined when compiled normally. This means we can not set breakpoints
in that function or step into it:

```
(dlv) b rand.RandomNumber
Command failed: location "rand.RandomNumber" not found
(dlv) s
4
> main.main() ./main.go:13 (PC: 0x45dca5)
Warning: debugging optimized function
Warning: listing may not match stale executable
     8:
     9: func main() {
    10:         runtime.Breakpoint()
    11:
    12:         println(rand.RandomNumber())
=>  13: }
```

You can compile your program with the argument `-gcflags=all='-N -l'` to
disable inlining and optimizations. This is exactly what Delve does for you
when you run `dlv debug` or `dlv test`:

```
$ go build -gcflags=all='-N -l'
$ dlv exec ./01-starting-a-debugging-session
(dlv) b rand.RandomNumber
Breakpoint 1 set at 0x467840 for github.com/jasonkeene/debugging-workshop/exercises/01-starting-a-debugging-session/rand.RandomNumber() ./rand/rand.go:3
(dlv)
```

### `attach`

If your program is already running, you can debug it by telling Delve to
attach to the PID. For this example, we will need to open three different
terminals. In the first terminal let's compile and run an HTTP server that
will return random numbers.

```
$ cd exercises/01-starting-a-debugging-session/server
$ go build -gcflags=all='-N -l'
$ ./server
listening on localhost:12345
PID is 94244
```

This server outputs its PID for us, so helpful! Let's copy this PID for
later. In the second terminal let's use `curl` to make a request:

```
$ curl localhost:12345
4
```

Awesome! It looks like we got our random number, and the server will have
output the message:

```
providing a random number
```

Excellent! Now in the third terminal, we will have Delve attach to this running
process:

```
$ dlv attach 94244
Type 'help' for list of commands.
(dlv)
```

This will suspend execution of the server and land us at a `(dlv)` prompt. To
debug our server we should set a breakpoint on the handler:

```
(dlv) b main.handler
Breakpoint 1 set at 0x1234cf3 for main.handler() ./src/github.com/jasonkeene/debugging-workshop/exercises/01-starting-a-debugging-session/server/main.go:18
```

We'll then want to continue execution of the server so it can accept requests:

```
(dlv) c
```

Now, when a request is made, the breakpoint will be hit and the debugger will
take control. Let's try that. In the second terminal make another request:

```
$ curl localhost:12345
```

This should block as the breakpoint has been hit:

```
> main.handler() ./src/github.com/jasonkeene/debugging-workshop/exercises/01-starting-a-debugging-session/server/main.go:17 (hits goroutine(8):1 total:1) (PC: 0x131bcb8)
    12:         fmt.Println("listening on localhost:12345")
    13:         fmt.Printf("PID is %d\n", os.Getpid())
    14:         log.Fatal(http.ListenAndServe("localhost:12345", http.HandlerFunc(handler)))
    15: }
    16:
=>  17: func handler(rw http.ResponseWriter, req *http.Request) {
    18:         fmt.Println("providing a random number")
    19:         fmt.Fprintf(rw, "%d\n", rand.RandomNumber())
    20: }
```

We can then inspect the request:

```
(dlv) p req.Header["User-Agent"][0]
"curl/7.64.1
```

And continue to return the response to `curl`:

```
(dlv) c
```

To exit the debugger send it a SIGINT via Ctrl+C, which will land you back to
a `(dlv)` prompt. You will want to clear the breakpoints with `clearall` then
you can type `q` to quit. Say no to not kill the server process as we are
going to use it for the next exercise.

### `trace`

The `trace` command is a bit different than the rest in that it doesn't
provide you a full debugging prompt. Instead, it allows you to log out when
specific functions are called in your program. You provide a regular
expression that matches the symbol names of the functions you would like to
trace. Delve will then set tracepoints on these functions. Tracepoints are
breakpoints that immediately log what function name and arguments and then
continue execution.

Let's trace some functions in the server we ran in the previous exercise. We
can start with tracing the handler:

```
$ dlv trace -p 94244 main.handler
```

Now use `curl` to send requests to the server:

```
$ curl localhost:12345
```

Delve should output something like:

```
> goroutine(49): main.handler(net/http.ResponseWriter(*net/http.response) 0xc0001bd888, ("*net/http.Request")(0xc000198f00))
 => ()
```

Since the arguments are pointer types you see the memory address for the
value that was passed. Let's try tracing every method of `*http.Request` to
see what methods get called.

```
$ dlv trace -p 94244 '^net/http.\(\*Request\)'
```

When you make a request you should see some output like this:

```
> goroutine(11): net/http.(*Request).isH2Upgrade(("*net/http.Request")(0xc000076600)) => (false)
> goroutine(11): net/http.(*Request).isH2Upgrade(("*net/http.Request")(0xc000076600)) => (false)
> goroutine(11): net/http.(*Request).ProtoAtLeast(("*net/http.Request")(0xc000076600), 1, 1) => (true)
> goroutine(11): net/http.(*Request).wantsHttp10KeepAlive(("*net/http.Request")(0xc000076600)) => (false)
> goroutine(11): net/http.(*Request).wantsClose(("*net/http.Request")(0xc000076600)) => (false)
> goroutine(11): net/http.(*Request).expectsContinue(("*net/http.Request")(0xc000076600)) => (false)
> goroutine(11): net/http.(*Request).ProtoAtLeast(("*net/http.Request")(0xc000076600), 1, 1) => (true)
> goroutine(11): net/http.(*Request).ProtoAtLeast(("*net/http.Request")(0xc000076600), 1, 0) => (true)
> goroutine(11): net/http.(*Request).ProtoAtLeast(("*net/http.Request")(0xc000076600), 1, 1) => (true)
```

You could do the same for `*http.response`:

```
$ dlv trace -p 94244 '^net/http.\(\*response\)'
> goroutine(13): net/http.(*response).Write(("*net/http.response")(0xc00007e380), []uint8 len: 2, cap: 32, [...])
> goroutine(13): net/http.(*response).write(("*net/http.response")(0xc00007e380), 2, []uint8 len: 2, cap: 32, [...], "")
> goroutine(13): net/http.(*response).WriteHeader(("*net/http.response")(0xc00007e380), 200)
 => ()
> goroutine(13): net/http.(*response).bodyAllowed(("*net/http.response")(0xc00007e380)) => (true)
 => (2,error nil)
 => (2,error nil)
> goroutine(13): net/http.(*response).finishRequest(("*net/http.response")(0xc00007e380))
 => ()
> goroutine(13): net/http.(*response).shouldReuseConnection(("*net/http.response")(0xc00007e380))
> goroutine(13): net/http.(*response).bodyAllowed(("*net/http.response")(0xc00007e380)) => (true)
> goroutine(13): net/http.(*response).closedRequestBodyEarly(("*net/http.response")(0xc00007e380)) => (false)
 => (true)
```

This allows you to determine empirically what functions are called, with what
arguments, and in what order. Much quicker than adding log lines to your
source code! To find out what functions are in your binary that you could
trace you can attach to the process again and run the `funcs` command:

```
$ dlv attach 94244
Type 'help' for list of commands.
(dlv) funcs
```

This will output all functions that can be traced. You can filter this list
down by regex:

```
(dlv) funcs ^net/http
```

These function names might seem a bit weird as that is not how you would
refer to them in a Go program.

These are the symbol names of the compiled binary. Symbol names identify
important places in the binary like the start of a function. They are fully
qualified to prevent collisions. As such, symbols for library code will have
the entire import path which can be quite long. You can see the full symbol
table with `objdump`:

```
$ objdump -t server
```

Feel free to `quit` the debugger and kill the server process as we won't be
needing it anymore.
