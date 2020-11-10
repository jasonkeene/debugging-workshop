
## Automating Delve

In certain situations you might want to automate Delve. For instance, if you
want to gather information automatically about a running process while
minimizing the time the process is suspended. Luckily there are a few great
options for doing this sort of thing.

### Command Scripts

Command scripts are simply a file that contain the commands you would like
Delve to run. Delve will run these commands for you and provide you the output
that way you don't have to type them in manually. To demonstrate this let's
debug the server:

```
$ cd exercises/07-automating-delve
$ dlv debug
Type 'help' for list of commands.
(dlv)
```

Now we are going to run the command script in this directory. It contains the
following commands:

```
b main.handler
c
p req.URL.Path
clearall
c
```

We can run this script with the `source` command:

```
(dlv) source script.dlv
Breakpoint 1 set at 0x131c398 for main.handler() ./main.go:17
listening on localhost:12345
PID is 19819
```

The server is now running with the breakpoint installed. Let's use curl to
cause the breakpoint to be hit:

```
$ curl localhost:12345/this/was/the/path
handler response
```

Delve then outputs this information for us when the breakpoint is hit:

```
> main.handler() ./main.go:17 (hits goroutine(21):1 total:1) (PC: 0x131c398)
    12:         fmt.Println("listening on localhost:12345")
    13:         fmt.Printf("PID is %d\n", os.Getpid())
    14:         log.Fatal(http.ListenAndServe("localhost:12345", http.HandlerFunc(handler)))
    15: }
    16:
=>  17: func handler(rw http.ResponseWriter, req *http.Request) {
    18:         fmt.Println("handler ran")
    19:         time.Sleep(time.Second)
    20:         rw.Write([]byte("handler response\n"))
    21: }
"/this/was/the/path"
Breakpoint 1 cleared at 0x131c398 for main.handler() ./main.go:17
handler ran
```

Command scripts are useful but fairly limited. You can not write complex
programs, just basic series of commands. For instance, if you wanted to run
different commands based on the state of the process, this would not be
possible.

### Starlark Scripts

Starlark is a dialect of Python used by Delve as a scripting language. You
can find the specification for the language [here][starlark].

Starlark scripts allow you to access all the functionality of Delve but from
a general purpose scripting langauge which allows for a much higher level of
automation. Additionally, you can create your own custom commands that you
can then call from the `(dlv)` prompt. For example if we want to display the
lines that created each goroutine we could implement that like this:

```
def command_goroutine_start_line(args):
    "prints the line of source code that started each currently running goroutine"
    gs = goroutines().Goroutines
    for g in gs:
        line = read_file(g.StartLoc.File).splitlines()[g.StartLoc.Line-1].strip()
        print(g.ID, "\t", g.StartLoc.File + ":" + str(g.StartLoc.Line), "\t", line)

def main():
    dlv_command("config alias goroutine_start_line gsl")
```

Any Starlark function that starts with the `command_` prefix is available by
whatever follows that prefix. You can also provide a `main` function that
will get evaluated on load. In this case, we configure an alias for our
command.

To load this script into the debugger we save it to a file with a `.star`
extension and then load it with the `source` command:

```
$ dlv debug
Type 'help' for list of commands.
(dlv) gsl
Command failed: command not available
(dlv) source script.star
(dlv) gsl
(dlv) c
listening on localhost:12345
PID is 92841
```

Now let's put some load on our server to create some goroutines. In another
terminal run:

```
$ go get github.com/tsliwowicz/go-wrk
$ go-wrk -c 5 -d 999 -T 1250 http://localhost:1234
```

From the debugger terminal hit `Ctrl+c` to stop the process:

```
received SIGINT, stopping process (will not forward signal)
Stopped at: 0x7fff6dfc7766
=>no source available
(dlv)
```

You can also kill the `go-wrk` process now. We can now see where all the
goroutines were created by running the `gsl` command:

```
(dlv) gsl
1        /usr/local/go/src/runtime/proc.go:114   func main() {
2        /usr/local/go/src/runtime/proc.go:246   func forcegchelper() {
3        /usr/local/go/src/runtime/mgcsweep.go:156       func bgsweep(c chan int) {
4        /usr/local/go/src/runtime/mgcscavenge.go:252    func bgscavenge(c chan int) {
18       /usr/local/go/src/runtime/mfinal.go:161         func runfinq() {
106      /usr/local/go/src/net/http/server.go:1794       func (c *conn) serve(ctx context.Context) {
58       /usr/local/go/src/net/http/server.go:689        func (cr *connReader) backgroundRead() {
105      /usr/local/go/src/net/http/server.go:1794       func (c *conn) serve(ctx context.Context) {
148      /usr/local/go/src/net/http/server.go:689        func (cr *connReader) backgroundRead() {
103      /usr/local/go/src/net/http/server.go:1794       func (c *conn) serve(ctx context.Context) {
104      /usr/local/go/src/net/http/server.go:1794       func (c *conn) serve(ctx context.Context) {
102      /usr/local/go/src/net/http/server.go:1794       func (c *conn) serve(ctx context.Context) {
164      /usr/local/go/src/net/http/server.go:689        func (cr *connReader) backgroundRead() {
140      /usr/local/go/src/net/http/server.go:689        func (cr *connReader) backgroundRead() {
139      /usr/local/go/src/net/http/server.go:689        func (cr *connReader) backgroundRead() {
```

It looks like each of the 5 concurrent requests we created with `go-wrk` were
serviced by two different goroutines, `backgroundRead` and `serve`.

At any time you can jump into a Starlark session from inside Delve by
providing `-` to the `source` command.

```
(dlv) source -
>>> [x*x for x in range(10)]
[0, 1, 4, 9, 16, 25, 36, 49, 64, 81]
```

Starlark support allows you to automate away the tedium of complex debugging
tasks. If you are interested in learning more see [the docs][starlark-docs].

### Delve's API

Delve's functionality is accessible via a JSON RPC API. This is the same
API that we used to remote debug our process earlier. JSON RPC is a fairly
straight forward protocol. You just open a TCP connection to the server and
then send newline delimited JSON objects and read the responses from the
connection.

Let's start the debugger, this time in headless mode:

```
$ dlv debug --headless --listen localhost:54321 --api-version 2 --accept-multiclient
API server listening at: 127.0.0.1:54321
debugserver-@(#)PROGRAM:LLDB  PROJECT:lldb-1100.0.30..1
 for x86_64.
Got a connection, launched process /Users/jasonkeene/src/github.com/jasonkeene/debugging-workshop/exercises/07-automating-delve/__debug_bin (pid = 87295).
```

We can initiate an RPC session by using netcat:

```
$ nc localhost 54321 | jq
```

We can then make an RPC by sending JSON objects:

```json
{"method":"RPCServer.FindLocation","params":[{"Loc":"main.handler"}]}
{
  "id": null,
  "result": {
    "Locations": [
      {
        "pc": 20038552,
        "file": "/Users/jasonkeene/src/github.com/jasonkeene/debugging-workshop/exercises/07-automating-delve/main.go",
        "line": 19,
        "function": {
          "name": "main.handler",
          "value": 20038528,
          "type": 0,
          "goType": 0,
          "optimized": false
        },
        "pcs": [
          20038552
        ]
      }
    ]
  },
  "error": null
}
```

You specify the method you would like to call and the parameters for that
method in a JSON object. That's pretty much it. A list of methods and what
arguments to pass are available on the [rpc2.RPCServer][rpcserver] type. If
you are interested in more details about JSON RPC itself the spec is located
at [jsonrpc.org][jsonrpc].

Let's set a breakpoint via the API. First, we need to get the address or
"program counter" where we want to set the breakpoint. The output of the RPC
contains this information under `result.Locations[0].pc`.

With that address we can now set a breakpoint. I will use the `Variables`
field to specify a variable we're interested in reading:

```json
{"method":"RPCServer.CreateBreakpoint","params":[{"Breakpoint":{"addr":20038552,"Variables":["someVar"]}}]}
{
  "id": null,
  "result": {
    "Breakpoint": {
      "id": 2,
      "name": "",
      "addr": 20038552,
      "addrs": [
        20038552
      ],
      "file": "/Users/jasonkeene/src/github.com/jasonkeene/debugging-workshop/exercises/07-automating-delve/main.go",
      "line": 19,
      "functionName": "main.handler",
      "Cond": "",
      "continue": false,
      "traceReturn": false,
      "goroutine": false,
      "stacktrace": 0,
      "LoadArgs": null,
      "LoadLocals": null,
      "hitCount": {},
      "totalHitCount": 0
    }
  },
  "error": null
}
```

We then need to continue the process:

```json
{"method":"RPCServer.Command","params":[{"Name":"continue"}]}
```

We can then trigger the breakpoint by using curl in another terminal:

```
$ curl localhost:12345
```

In our RPC terminal you should see a bunch of output when the breakpoint is
hit. Inside that output you will find a section for the variables we were
interested in inspecting:

```json
{
  "result": {
    "State": {
      "currentThread": {
        "breakPointInfo": {
          "variables": [
            {
              "name": "someVar",
              "addr": 22073568,
              "onlyAddr": false,
              "type": "string",
              "realType": "string",
              "flags": 0,
              "kind": 24,
              "value": "some variable",
              "len": 13,
              "cap": 0,
              "children": [],
              "base": 20507671,
              "unreadable": "",
              "LocationExpr": "[block] DW_OP_addr 0x150d0e0 ",
              "DeclLine": 0
            }
          ]
        }
      }
    }
  }
}
```

### API Go Client

Now that we are familiar with the JSON RPC API let's use the Go client to
automate debugging a program. In the `debugme/ring-buffer` directory you will
find a ring buffer program. This ring buffer has a goroutine that writes
sequential integers in order. The main goroutine then reads these integers
and checks to ensure they are in order. If you run this program it appears to
run fine for a while but then you will start getting errors:

```
$ cd debugme/ring-buffer
$ go run .
.............................................................................................................................................................................................................................................................................................................
WARNING: data is not in order or data is missing
```

Each dot represents the entire ring buffer capacity being written to and then
read once. It is strange that the warning is intermittent. Sometimes it will
just go away after a while.

Let's inspect the read and write indices of the buffer using a Delve and a Go
program. There is a program in the `controller` directory that uses the Delve
API to read these indices. First thing we need to do is start Delve in
headless mode:

```go
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
```

Once Delve is started we need to wait for it to be up and accepting
connections. We can just attempt to dial multiple times wtih a timeout:

```go
for {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err == nil {
		conn.Close()
		return
	}
}
```

We can then create our client and tell Delve to continue to start the process:

```go
client := rpc2.NewClient(addr)
client.Continue()
defer func() {
	client.Halt()
	client.Detach(false)
}()
```

Now, what we want to do is monitor the read and write indices while our
process is running to see what happens to them when we see that warning. An
easy way to do this is to wake up every second, install a breakpoint, read
the indices, output them to the user, clear the breakpoint, and then go back
to sleep.

First, let's create a breakpoint object we will reuse to both create and clear
the breakpoint. We need to add the breakpoint somewhere that is being
executed frequently so it will be hit quickly. There are many places in the
program we could do this, the `read` method seems like a fine place. We will
need to find the location in memory for this method.

```go
locations, _ := client.FindLocation(
	api.EvalScope{
		GoroutineID: -1,
	},
	"main.(*ringBuffer).read",
	true,
)
pc := locations[0].PC
```

Now we can create our breakpoint object:

```go
bp := &api.Breakpoint{
	Name: "readIndices",
	Addr: pc,
	Variables: []string{
		"b.writeIndex",
		"b.readIndex",
	},
}
```

Finally, we loop forever reading the indices and reporting them to the user:

```go
for {
	client.Halt()
	client.CreateBreakpoint(bp)
	state := <-client.Continue()
	client.ClearBreakpointByName(bp.Name)
	client.Continue()

	writeIndex, _ := strconv.Atoi(state.CurrentThread.BreakpointInfo.Variables[0].Value)
	readIndex, _ := strconv.Atoi(state.CurrentThread.BreakpointInfo.Variables[1].Value)

	fmt.Printf("\nw: %d r: %d delta: %d\n", writeIndex, readIndex, writeIndex-readIndex)

	time.Sleep(time.Second)
}
```

Let's run this controller program to see what is going on with our indices:

```
$ cd exercises/07-automating-delve/controller
$ go run .
API server listening at: 127.0.0.1:12345

w: 0 r: 0 delta: 0
....................................
w: 705120 r: 704672 delta: 448
...................................
w: 1420897 r: 1419224 delta: 1673
....................................
w: 2132082 r: 2131801 delta: 281
...................................
w: 2824359 r: 2824173 delta: 186
...................................
w: 3529550 r: 3529445 delta: 105
...................................
w: 4234169 r: 4232831 delta: 1338
..................................
w: 4913061 r: 4911098 delta: 1963
```

After a bit of time we will see the warnings again:

```
WARNING: data is not in order or data is missing
.........
w: 22107461 r: 22087460 delta: 20001
.......WARNING: data is not in order or data is missing
```

### DEBUGME: Ring Buffer

Interesting that the delta of the indices when the warning is occurring is
`20001` 🤔 Using this new bit of information can you solve this bug?

This is just a basic example of what is possible when it comes to automating
Delve. The JSON RPC API has a lot more functionality than we have covered
here.

[starlark]: https://github.com/google/starlark-go/blob/master/doc/spec.md
[starlark-docs]: https://github.com/go-delve/delve/blob/master/Documentation/cli/starlark.md
[rpcserver]: https://godoc.org/github.com/go-delve/delve/service/rpc2#RPCServer
[jsonrpc]: https://www.jsonrpc.org/
