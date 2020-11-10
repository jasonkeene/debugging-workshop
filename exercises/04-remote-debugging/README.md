
## Remote Debugging

Delve has the ability to attach to processes over a network via a JSON RPC
API. You attach to your target process as a server and then connect that
server with a client:

```
      You
       +
       |
       |
+------v-------+                    +--------------+
|              |                    |              |
| Delve Client | <--- JSON RPC ---> | Delve Server |
|              |                    |              |
+--------------+                    +------+-------+
                                           |
                                           |
                                   +-------v--------+
                                   |                |
                                   | Target Process |
                                   |                |
                                   +----------------+
```

Let's first attach Delve to the process locally in headless mode:

```
$ cd exercises/04-remote-debugging
$ dlv debug --headless --listen localhost:12345 --api-version 2 --accept-multiclient
API server listening at: 127.0.0.1:12345
debugserver-@(#)PROGRAM:LLDB  PROJECT:lldb-1100.0.30..1
 for x86_64.
Got a connection, launched process /Users/jasonkeene/src/github.com/jasonkeene/debugging-workshop/exercises/04-remote-debugging/__debug_bin (pid = 25685).
```

We just told Delve to start in headless mode and listen on `localhost:12345`
for new connections. Delve can listen on any interface/port you'd like to be
able to connect from.

⚠️ A bit of a security warning here: do **NOT** allow Delve to be accessed
across an insecure network. Delve performs no authentication or authorization
on client commands. Anyone who has access to Delve in this way has full
control over your process and can read and write to memory and execute code.

With that warning out of the way let's connect a Delve client to the Delve
server we have safely listening on `localhost`. In another terminal:

```
$ dlv connect localhost:12345 --api-version 2
Type 'help' for list of commands.
(dlv)
```

Now we can drive the debugger using the normal commands as if we were locally
attached to the process. Let's just continue for now:

```
(dlv) c
```

### Reading Remote Memory

The process we are debugging is an encryption oracle that will take data from
a request and encrypt it with a secret key. We can see how that works by
using `curl` in another terminal:

```
curl -s -d "foobar" localhost:54321 | base64
```

Let's try and get access to that secret key. For starters, we'll need to set a
breakpoint. We can use Ctrl+C to send SIGINT to the Delve client to halt the
process. After sending the signal you will be prompted if you want to pause
the target or quit the client. We want to pause:

```
Would you like to [p]ause the target (returning to Delve's prompt) or [q]uit this client (leaving the target running) [p/q]? p
Stopped at: 0x7fff72b7b882
```

Let's set a breakpoint on our handler and continue:

```
(dlv) b main.handler
Breakpoint 1 set at 0x15532b8 for main.handler() ./src/github.com/jasonkeene/debugging-workshop/exercises/04-remote-debugging/main.go:32
(dlv) c
```

Now run our `curl` command again in another terminal:

```
curl -s -d "foobar" localhost:54321 | base64
```

Once the breakpoint is hit we can run:

```
(dlv) p a.ps.Primary.Primitive.Key
[]uint8 len: 32, cap: 32, [30,192,15,62,147,199,99,14,162,5,38,143,15,24,95,79,93,230,107,112,35,181,10,125,7,183,153,8,38,205,17,67]
```

and there is our private key! Imagine any actor on our network being able to
perform such an action. Hopefully, this underscores the power and potential
danger of remote debugging.
