
## Navigating Your Program

Once you are in a debugging session you need to be able to control the
execution of the program. Let's see how to do that. First launch the debugger
in this directory:

```
$ cd exercises/02-navigating-your-program
$ dlv debug
Type 'help' for list of commands.
(dlv)
```

You should also load up `main.go` in your editor. This program starts an HTTP
server with a handler and sends an HTTP request to it then exits. Fairly
straight forward but this gives us some code to explore.

First, set a breakpoint at `main.handler` so we know when it is invoked.

```
(dlv) b main.handler
Breakpoint 1 set at 0x1389998 for main.handler() ./main.go:36
```

and continue until it gets hit:

```
(dlv) c
> main.handler() ./main.go:36 (hits goroutine(5):1 total:1) (PC: 0x1389998)
    31:         }()
    32:
    33:         return l.Addr().String(), nil
    34: }
    35:
=>  36: func handler(rw http.ResponseWriter, r *http.Request) {
    37:         data, err := json.Marshal("request handled")
    38:         if err != nil {
    39:                 rw.WriteHeader(http.StatusInternalServerError)
    40:                 return
    41:         }
```

### Inspecting State

The first thing we might want to do is see how we got to this handler. We can
list out a stackrace by running the `stack` or `bt` command:

```
(dlv) bt
0  0x0000000001389998 in main.handler
   at ./main.go:36
1  0x000000000134b2a4 in net/http.HandlerFunc.ServeHTTP
   at /usr/local/go/src/net/http/server.go:2042
2  0x000000000134ebcb in net/http.serverHandler.ServeHTTP
   at /usr/local/go/src/net/http/server.go:2843
3  0x000000000134a585 in net/http.(*conn).serve
   at /usr/local/go/src/net/http/server.go:1925
4  0x00000000010762e1 in runtime.goexit
   at /usr/local/go/src/runtime/asm_amd64.s:1374
```

This lists out every frame starting at the current frame where the breakpoint
was hit (0) and going all the way back to the goroutine's top frame (4 in
this case). We can navigate up and down the stack by using the `up` and
`down` commands:

```
(dlv) up
> main.handler() ./main.go:36 (hits goroutine(5):1 total:1) (PC: 0x1389998)
Frame 1: /usr/local/go/src/net/http/server.go:2042 (PC: 134b2a4)
  2037: // Handler that calls f.
  2038: type HandlerFunc func(ResponseWriter, *Request)
  2039:
  2040: // ServeHTTP calls f(w, r).
  2041: func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
=>2042:         f(w, r)
  2043: }
  2044:
  2045: // Helper handlers
  2046:
  2047: // Error replies to the request with the specified error message and HTTP code.
```

This shows the code for the method that called `main.handler`. In this case
it is the `HandlerFunc`'s `ServeHTTP` method which isn't terribly
interesting. Let's go up one more:

```
(dlv) up
> main.handler() ./main.go:36 (hits goroutine(5):1 total:1) (PC: 0x1389998)
Frame 2: /usr/local/go/src/net/http/server.go:2843 (PC: 134ebcb)
  2838:                 handler = DefaultServeMux
  2839:         }
  2840:         if req.RequestURI == "*" && req.Method == "OPTIONS" {
  2841:                 handler = globalOptionsHandler{}
  2842:         }
=>2843:         handler.ServeHTTP(rw, req)
  2844: }
  2845:
  2846: // ListenAndServe listens on the TCP network address srv.Addr and then
  2847: // calls Serve to handle requests on incoming connections.
  2848: // Accepted connections are configured to enable TCP keep-alives.
```

Cool, now we are in the frame for `serverHandler.ServeHTTP`. We can list the
arguments and local variables for this method:

```
(dlv) args
sh = net/http.serverHandler {srv: ("*net/http.Server")(0xc000072000)}
rw = net/http.ResponseWriter(*net/http.response) 0xc00009f900
req = ("*net/http.Request")(0xc000100100)
(dlv) locals
handler = net/http.Handler(net/http.HandlerFunc) main.handler
```

At any point you can use the `list` or `ls` command to print out the context
of where you are:

```
(dlv) ls
Goroutine 5 frame 2 at /usr/local/go/src/net/http/server.go:2843 (PC: 0x134ebcb)
  2838:                 handler = DefaultServeMux
  2839:         }
  2840:         if req.RequestURI == "*" && req.Method == "OPTIONS" {
  2841:                 handler = globalOptionsHandler{}
  2842:         }
=>2843:         handler.ServeHTTP(rw, req)
  2844: }
  2845:
  2846: // ListenAndServe listens on the TCP network address srv.Addr and then
  2847: // calls Serve to handle requests on incoming connections.
  2848: // Accepted connections are configured to enable TCP keep-alives.
```

If you want to set the current frame back to where the breakpoint was hit you
can do:

```
(dlv) frame 0
> main.handler() ./main.go:36 (hits goroutine(5):1 total:1) (PC: 0x1389998)
Frame 0: ./main.go:36 (PC: 1389998)
    31:         }()
    32:
    33:         return l.Addr().String(), nil
    34: }
    35:
=>  36: func handler(rw http.ResponseWriter, r *http.Request) {
    37:         data, err := json.Marshal("request handled")
    38:         if err != nil {
    39:                 rw.WriteHeader(http.StatusInternalServerError)
    40:                 return
    41:         }
```

Using these commands we can inspect the state for the goroutine where the
breakpoint was hit. What if we wanted to inspect the state of a different
goroutine? Not surprisingly Delve has first-class support for goroutines! We
can use the `goroutines` command to list out the current goroutines:

```
(dlv) goroutines
  Goroutine 1 - User: /usr/local/go/src/net/http/transport.go:2565 net/http.(*persistConn).roundTrip (0x1371545)
  Goroutine 2 - User: /usr/local/go/src/runtime/proc.go:307 runtime.gopark (0x1041135)
  Goroutine 3 - User: /usr/local/go/src/runtime/proc.go:307 runtime.gopark (0x1041135)
  Goroutine 4 - User: /usr/local/go/src/runtime/proc.go:307 runtime.gopark (0x1041135)
* Goroutine 5 - User: ./main.go:36 main.handler (0x1389998) (thread 9113720)
  Goroutine 18 - User: /usr/local/go/src/runtime/proc.go:307 runtime.gopark (0x1041135)
  Goroutine 21 - User: /usr/local/go/src/runtime/netpoll.go:220 internal/poll.runtime_pollWait (0x1070d45)
  Goroutine 24 - User: /usr/local/go/src/runtime/netpoll.go:220 internal/poll.runtime_pollWait (0x1070d45)
  Goroutine 25 - User: /usr/local/go/src/net/http/transport.go:2340 net/http.(*persistConn).writeLoop (0x137044a)
  Goroutine 34 - User: /usr/local/go/src/net/http/server.go:689 net/http.(*connReader).backgroundRead (0x13402a0)
```

You can then switch to a specific goroutine using the `goroutine` command:

```
(dlv) goroutine 1
Switched from 0 to 1 (thread 9113727)
(dlv) bt
 0  0x0000000001041135 in runtime.gopark
    at /usr/local/go/src/runtime/proc.go:307
 1  0x00000000010515ca in runtime.selectgo
    at /usr/local/go/src/runtime/select.go:319
 2  0x0000000001371545 in net/http.(*persistConn).roundTrip
    at /usr/local/go/src/net/http/transport.go:2565
 3  0x00000000013620a6 in net/http.(*Transport).roundTrip
    at /usr/local/go/src/net/http/transport.go:582
 4  0x000000000133e13e in net/http.(*Transport).RoundTrip
    at /usr/local/go/src/net/http/roundtrip.go:17
 5  0x00000000012e81b9 in net/http.send
    at /usr/local/go/src/net/http/client.go:252
 6  0x00000000012e7954 in net/http.(*Client).send
    at /usr/local/go/src/net/http/client.go:176
 7  0x00000000012eb60f in net/http.(*Client).do
    at /usr/local/go/src/net/http/client.go:718
 8  0x00000000012ea53e in net/http.(*Client).Do
    at /usr/local/go/src/net/http/client.go:586
 9  0x00000000012e9e6a in net/http.(*Client).Get
    at /usr/local/go/src/net/http/client.go:475
10  0x00000000012e9cad in net/http.Get
    at /usr/local/go/src/net/http/client.go:447
11  0x0000000001389c34 in main.sendRequest
    at ./main.go:47
12  0x000000000138975a in main.main
    at ./main.go:17
13  0x0000000001040d0f in runtime.main
    at /usr/local/go/src/runtime/proc.go:204
14  0x00000000010762e1 in runtime.goexit
    at /usr/local/go/src/runtime/asm_amd64.s:1374
```

Many of Delve's commands can contain a `goroutine` and `frame` prefix to run
that command in that goroutine and/or frame:

```
(dlv) help list
Show source code.

        [goroutine <n>] [frame <m>] list [<linespec>]

Show source around current point or provided linespec.

For example:

        frame 1 list 69
        list testvariables.go:10000
        list main.main:30
        list 40
```

For example, we can `list` frame 11 of this goroutine without having to
navigate to that frame:

```

(dlv) frame 11 ls
Goroutine 1 frame 11 at /Users/jasonkeene/src/github.com/jasonkeene/debugging-workshop/exercises/02-navigating-your-program/main.go:47 (PC: 0x1389c34)
    42:
    43:         rw.Write(data)
    44: }
    45:
    46: func sendRequest(addr string) {
=>  47:         _, err := http.Get(fmt.Sprintf("http://%s", addr))
    48:         if err != nil {
    49:                 log.Print(err)
    50:         }
    51: }
```

This shows the frame where the HTTP request was made. Let's reset real quick
to get back to our original breakpoint:

```
(dlv) r
Process restarted with PID 13497
(dlv) c
> main.handler() ./main.go:36 (hits goroutine(5):1 total:1) (PC: 0x1389998)
    31:         }()
    32:
    33:         return l.Addr().String(), nil
    34: }
    35:
=>  36: func handler(rw http.ResponseWriter, r *http.Request) {
    37:         data, err := json.Marshal("request handled")
    38:         if err != nil {
    39:                 rw.WriteHeader(http.StatusInternalServerError)
    40:                 return
    41:         }
```

### Controlling Execution

In addition to inspecting state of different goroutines we can also step
through the execution of our program. Let's use the `next` or `n` command to
go to the next line:

```
(dlv) n
> main.handler() ./main.go:37 (PC: 0x13899af)
    32:
    33:         return l.Addr().String(), nil
    34: }
    35:
    36: func handler(rw http.ResponseWriter, r *http.Request) {
=>  37:         data, err := json.Marshal("request handled")
    38:         if err != nil {
    39:                 rw.WriteHeader(http.StatusInternalServerError)
    40:                 return
    41:         }
```

We can use `next` to step over each line and allow it to execute. If we try
to print `data` from here we can see it hasn't been declared yet:

```
(dlv) p data
Command failed: could not find symbol value for data
```

Let's step over the `json.Marshal` call to see what it assigns into `data`:

```
(dlv) n
> main.handler() ./main.go:38 (PC: 0x1389a4f)
    33:         return l.Addr().String(), nil
    34: }
    35:
    36: func handler(rw http.ResponseWriter, r *http.Request) {
    37:         data, err := json.Marshal("request handled")
=>  38:         if err != nil {
    39:                 rw.WriteHeader(http.StatusInternalServerError)
    40:                 return
    41:         }
    42:
    43:         rw.Write(data)
(dlv) p data
[]uint8 len: 17, cap: 32, [34,114,101,113,117,101,115,116,32,104,97,110,100,108,101,100,34]
(dlv) p string(data)
"\"request handled\""
```

Pretty much what we'd expect. `err` was set to `nil`. I wonder if we can
force the execution down this error branch by setting error 🤔

```
(dlv) p err
error nil
(dlv) call err = errors.New("😈")
> main.handler() ./main.go:38 (PC: 0x1389a4f)
    33:         return l.Addr().String(), nil
    34: }
    35:
    36: func handler(rw http.ResponseWriter, r *http.Request) {
    37:         data, err := json.Marshal("request handled")
=>  38:         if err != nil {
    39:                 rw.WriteHeader(http.StatusInternalServerError)
    40:                 return
    41:         }
    42:
    43:         rw.Write(data)
(dlv) p err
error(*errors.errorString) *{s: "😈"}
(dlv) n
> main.handler() ./main.go:39 (PC: 0x1389a59)
    34: }
    35:
    36: func handler(rw http.ResponseWriter, r *http.Request) {
    37:         data, err := json.Marshal("request handled")
    38:         if err != nil {
=>  39:                 rw.WriteHeader(http.StatusInternalServerError)
    40:                 return
    41:         }
    42:
    43:         rw.Write(data)
    44: }
```

Hah! That's quite mischievous! Manipulating memory like this is useful when
exploring code. You can do the same with the `set` command if `call` doesn't
work for you. However, `set` requires that your error value already exists in
memory somewhere. For example:

```
(dlv) set err = io.EOF
(dlv) p err
error(*errors.errorString) *{s: "EOF"}
```

Now that we forced this error branch to be executed how about we step into
the `WriteHeader` method. We can do that with the `step` or `s` command:

```
(dlv) s
> net/http.(*response).WriteHeader() /usr/local/go/src/net/http/server.go:1133 (PC: 0x13441fb)
  1128:                 }
  1129:         }
  1130:         return frame
  1131: }
  1132:
=>1133: func (w *response) WriteHeader(code int) {
  1134:         if w.conn.hijacked() {
  1135:                 caller := relevantCaller()
  1136:                 w.conn.server.logf("http: response.WriteHeader on hijacked connection from %s (%s:%d)", caller.Function, path.Base(caller.File), caller.Line)
  1137:                 return
  1138:         }
```

The `WriteHeader` method writes the HTTP header out to the connection,
including things like the status code. We can skip a few lines down by
providing `next` an argument which is the number of lines to step over:

```
(dlv) n 6
> net/http.(*response).WriteHeader() /usr/local/go/src/net/http/server.go:1148 (PC: 0x134471f)
  1143:         }
  1144:         checkWriteHeaderCode(code)
  1145:         w.wroteHeader = true
  1146:         w.status = code
  1147:
=>1148:         if w.calledHeader && w.cw.header == nil {
  1149:                 w.cw.header = w.handlerHeader.Clone()
  1150:         }
  1151:
  1152:         if cl := w.handlerHeader.get("Content-Length"); cl != "" {
  1153:                 v, err := strconv.ParseInt(cl, 10, 64)
```

It looks like we are right after the line where we set the status code. Let's
monkey with that a bit for fun:

```
> net/http.(*response).WriteHeader() /usr/local/go/src/net/http/server.go:1148 (PC: 0x134471f)
  1143:         }
  1144:         checkWriteHeaderCode(code)
  1145:         w.wroteHeader = true
  1146:         w.status = code
  1147:
=>1148:         if w.calledHeader && w.cw.header == nil {
  1149:                 w.cw.header = w.handlerHeader.Clone()
  1150:         }
  1151:
  1152:         if cl := w.handlerHeader.get("Content-Length"); cl != "" {
  1153:                 v, err := strconv.ParseInt(cl, 10, 64)
(dlv) p w.status
500
(dlv) set w.status = 503
(dlv) p w.status
503
```

We can then step out of this function by using the `stepout` or `so`. This
continues execution of the function and when the function returns Delve halts
execution again:

```
(dlv) so
> main.handler() ./main.go:41 (PC: 0x138a09e)
Values returned:

    36:
    37: func handler(rw http.ResponseWriter, r *http.Request) {
    38:         data, err := json.Marshal("request handled")
    39:         if err != nil {
    40:                 rw.WriteHeader(http.StatusInternalServerError)
=>  41:                 return
    42:         }
    43:
    44:         rw.Write(data)
    45: }
    46:
```

If we `continue` the program now we should see our mutated status code
returned to the client:

```
(dlv) c
resp: 503
Process 15092 has exited with status 0
```

### Breakpoints

Breakpoints can not only be set by line number or function name as we've
seen. You can set a breakpoint with a location specifier. These are all
valid:

- `<line>` Specifies the line in the current file
- `<function>[:<line>]` Specifies the line inside function
- `<filename>:<line>` Specifies the line in filename
- `/<regex>/` Specifies the location of all the functions matching regex
- `*<address>` Specifies the location of memory address address
- `+<offset>` Specifies the line offset lines after the current one
- `-<offset>` Specifies the line offset lines before the current one

You can also give your breakpoint a name to help you refer to it later:

```
$ dlv debug
(dlv) b handler *0x789238
Breakpoint handler set at 0x789238 for main.handler() ./main.go:39
(dlv) help clear
Deletes breakpoint.

        clear <breakpoint name or id>
(dlv) clear handler
Breakpoint handler cleared at 0x789238 for main.handler() ./main.go:39
```

Sometimes you want to set a breakpoint but don't want to break every time it
is hit, only on a certain condition. Delve supports this with the `condtion`
or `cond` command. Let's see how that works:

```
$ cd exercises/02-navigating-your-program/condition# dlv debug
(dlv) b handler main.handler
Breakpoint handler set at 0x7306b8 for main.handler() ./main.go:17
(dlv) cond handler r.URL.Path == "/break"
(dlv) c
listening on localhost:12345
PID is 25348
```

Now when you curl the server it will respond without hitting the breakpoint:

```
$ curl localhost:12345
"request handled"
```

However, if you curl it with the right path it will break:

```
$ curl localhost:12345/break
```

```
> [handler] main.handler() ./main.go:17 (hits goroutine(23):1 total:1) (PC: 0x7306b8)
    12:         fmt.Println("listening on localhost:12345")
    13:         fmt.Printf("PID is %d\n", os.Getpid())
    14:         log.Fatal(http.ListenAndServe("localhost:12345", http.HandlerFunc(handler)))
    15: }
    16:
=>  17: func handler(rw http.ResponseWriter, r *http.Request) {
    18:         data, err := json.Marshal("request handled")
    19:         if err != nil {
    20:                 rw.WriteHeader(http.StatusInternalServerError)
    21:                 return
    22:         }
(dlv)
```

This is incredibly useful in eliminating situations where most of the time
you do not want to break. Note, the implementation of this is that the
breakpoint is still hit but the process is just continued if the condition
evaluates to false.

With the `on` command you can register commands to run after the breakpoint
fires. This is useful for gathering information. For example:

```
(dlv) on handler p r.Header["User-Agent"][0]
(dlv) c
```

```
$ curl localhost:12345/break
```

```
> [handler] main.handler() ./main.go:17 (hits goroutine(11):1 total:1) (PC: 0x7306b8)
        r.Header["User-Agent"][0]: "curl/7.58.0"
    12:         fmt.Println("listening on localhost:12345")
    13:         fmt.Printf("PID is %d\n", os.Getpid())
    14:         log.Fatal(http.ListenAndServe("localhost:12345", http.HandlerFunc(handler)))
    15: }
    16:
=>  17: func handler(rw http.ResponseWriter, r *http.Request) {
    18:         data, err := json.Marshal("request handled")
    19:         if err != nil {
    20:                 rw.WriteHeader(http.StatusInternalServerError)
    21:                 return
    22:         }
(dlv)
```

That is about all you need to begin navigating your program.

### DEBUGME: Square Root

In the `debugme/sqrt` directory there is a program that is supposed to allow
you to take the square root of a negative number. For some reason, it crashes
though. Can you figure out why?

### DEBUGME: Goroutine Leak 1

In the `debugme/goroutine-leak-1` directory there is a program that is very
similar to the one we have been debugging. Try your hand at debugging it!

### DEBUGME:Deadlock

In the `debugme/deadlock` directory there is a program that runs correctly
for a while but then deadlocks. See if you can solve it!
