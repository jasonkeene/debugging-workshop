
## Debugging Inside a Container

You might often find yourself needing to debug a process running inside
a container. Perhaps your local development setup requires running multiple
containers with docker-compose or you are trying to debug in a live
environment. Debugging inside a container introduces a few difficulties that
we will go over.

First, let's build and run the container we are going to be using for this
exercise:

```
$ cd exercises/05-debugging-inside-a-container/
$ docker build -t debugme .
Sending build context to Docker daemon  4.096kB
Step 1/8 : FROM golang:1.15 AS builder
 ---> 4a581cd6feb1
Step 2/8 : WORKDIR /
 ---> Using cache
 ---> 8c7b40818290
Step 3/8 : COPY main.go /
 ---> 721415a68342
Step 4/8 : RUN go build -o debugme main.go
 ---> Running in 40c516e58b62
Removing intermediate container 40c516e58b62
 ---> cac05d621585
Step 5/8 : FROM golang:1.15
 ---> 4a581cd6feb1
Step 6/8 : COPY --from=builder /debugme /
 ---> d76a669d30a8
Step 7/8 : EXPOSE 12345/tcp
 ---> Running in 98d33ade9f14
Removing intermediate container 98d33ade9f14
 ---> d70e460ee52e
Step 8/8 : CMD ["/debugme"]
 ---> Running in e07393f20d60
Removing intermediate container e07393f20d60
 ---> 67678df96a15
Successfully built 67678df96a15
Successfully tagged debugme:latest
$ docker run -it --rm -p 12345:12345 --name debugme debugme
listening on :12345
```

Now in another terminal, we should be able to curl this server:

```
$ curl localhost:12345
handler ran
```

We will need a version of the `dlv` command inside the container to exec.
This version of Delve needs to be built for Linux:

```
$ GOOS=linux go build github.com/go-delve/delve/cmd/dlv
$ docker cp ./dlv debugme:/dlv
```

Now that we have the `dlv` command inside our container we should be able to
just exec in and attach to the process:

```
$ docker exec -it debugme /bin/bash
```

Containers typically utilize a PID namespace so the process we are normally
interested in attaching to will be PID 1:

```
root@4c74243217d6:/go# ps aux
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.0  0.2 1077000 4924 pts/0    Ssl+ 01:10   0:00 /debugme
root        13  0.3  0.1   3868  3184 pts/1    Ss   01:14   0:00 /bin/bash
root        19  0.0  0.1   7640  2696 pts/1    R+   01:14   0:00 ps aux
```

Let's try that:

```
root@4c74243217d6:/go# /dlv attach 1
Could not attach to pid 1: this could be caused by a kernel security setting, try writing "0" to /proc/sys/kernel/yama/ptrace_scope
```

Well, that didn't work. In most environments you will likely have to change
the `ptrace_scope` yama setting. To do this you'll need to either be `root`
or have `CAP_SYS_PTRACE`. You might think you have root because of the
prompt, however when you try to set this setting you get an error:

```
root@4c74243217d6:/go# cat /proc/sys/kernel/yama/ptrace_scope
1
root@4c74243217d6:/go# echo 0 > /proc/sys/kernel/yama/ptrace_scope
bash: /proc/sys/kernel/yama/ptrace_scope: Read-only file system
```

This is because you are only root in the user namespace for the container,
not real root. An easy way to get root shell on the docker host is to use
`nsenter`:

```
$ docker run --rm -it --privileged --pid=host debian nsenter -t 1 -m -u -n -i sh
```

And then disable this restriction:

```
/ # cat /proc/sys/kernel/yama/ptrace_scope
1
/ # echo 0 > /proc/sys/kernel/yama/ptrace_scope
/ # cat /proc/sys/kernel/yama/ptrace_scope
0
```

This will only work if you can run privileged containers. Now that yama is
happy, let's try to attach again:

```
root@4c74243217d6:/go# /dlv attach 1
Type 'help' for list of commands.
(dlv)
```

Cool, we are attached. Let's set a breakpoint on the handler:

```
(dlv) b main.handler
Breakpoint 1 set at 0x634613 for main.handler() /main.go:13
(dlv) c
```

Now let's curl the server again to trigger the breakpoint:

```
$ curl localhost:12345
```

```
> main.handler() /main.go:13 (hits goroutine(19):1 total:1) (PC: 0x634613)
Warning: debugging optimized function
(dlv) ls
> main.handler() /main.go:13 (hits goroutine(19):1 total:1) (PC: 0x634613)
Warning: debugging optimized function
Command failed: open /main.go: no such file or directory
(dlv)
```

The breakpoint fired but we are unable to see the source code. Additionally,
we get warnings about debugging optimized functions. Most container binaries
will have optimizations and inlining enabled. Delve will do it's best to
allow you to debug the optimized binary but you might get unexpected results.

We can see the function names in the stack trace as well as file and line
numbers for where those functions are defined:

```
(dlv) bt
0  0x0000000000634613 in main.handler
   at /main.go:13
1  0x0000000000621ba4 in net/http.HandlerFunc.ServeHTTP
   at /usr/local./src/net/http/server.go:2042
2  0x00000000006242e3 in net/http.serverHandler.ServeHTTP
   at /usr/local./src/net/http/server.go:2843
3  0x0000000000620bed in net/http.(*conn).serve
   at /usr/local./src/net/http/server.go:1925
4  0x000000000046b041 in runtime.goexit
   at /usr/local./src/runtime/asm_amd64.s:1374
```

This information is contained in the binary. It is usually enough to
cross-reference with source code, however, if we want to debug with source
code we just need to make it available to Delve. Let's copy our source into
the container:

```
$ docker cp main.go debugme:/
```

Now when we run `ls` we can see source code:

```
> main.handler() /main.go:13 (hits goroutine(19):1 total:1) (PC: 0x634613)
Warning: debugging optimized function
     8: func main() {
     9:         fmt.Println("listening on :12345")
    10:         log.Fatal(http.ListenAndServe(":12345", http.HandlerFunc(handler)))
    11: }
    12:
=>  13: func handler(rw http.ResponseWriter, req *http.Request) {
    14:         fmt.Println("handler ran")
    15:         rw.Write([]byte("handler ran\n"))
    16: }
```

If you copy the source code to a different path than where it was originally
built you will need to tell Delve where to find the code. You can do that
with the `substitute-path` setting:

```
(dlv) config substitute-path /original/build/path /new/path
```

## Scratch Images

Some folks like to build their docker images with no files other than the
fully statically linked Go binary. Let's build a docker image like that:

```
$ docker build -f Dockerfile.scratch -t debugme-scratch .
Sending build context to Docker daemon  15.91MB
Step 1/8 : FROM golang:1.15 AS builder
 ---> ad12afc2d6d6
Step 2/8 : WORKDIR /
 ---> Running in d1687c0dcae9
Removing intermediate container d1687c0dcae9
 ---> bd061ae6a047
Step 3/8 : COPY main.go /
 ---> 5ac405101e27
Step 4/8 : RUN CGO_ENABLED=0 go build -o debugme main.go
 ---> Running in cddbbbc79ae0
Removing intermediate container cddbbbc79ae0
 ---> ecebe2044a1c
Step 5/8 : FROM scratch
 --->
Step 6/8 : COPY --from=builder /debugme /
 ---> 4d0a2092eda1
Step 7/8 : EXPOSE 12345/tcp
 ---> Running in 4115accf562a
Removing intermediate container 4115accf562a
 ---> f613d734989c
Step 8/8 : CMD ["/debugme"]
 ---> Running in 6e4096abb11e
Removing intermediate container 6e4096abb11e
 ---> 7d30257de775
Successfully built 7d30257de775
Successfully tagged debugme-scratch:latest
```

And run it as we did before:

```
$ docker run -it --rm -p 12345:12345 --name debugme-scratch debugme-scratc
listening on :12345
```

The first problem we will run into is that Delve dynamically links against
libc, vdso, and some other shared libraries by default:

```
$ ldd dlv
        linux-vdso.so.1 (0x00007fffbfe69000)
        libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007f7bb1484000)
        libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f7bb1093000)
        /lib64/ld-linux-x86-64.so.2 (0x00007f7bb16a3000)
```

We need to build Delve as a static binary with no dynamic
linking:

```
$ CGO_ENABLED=0 GOOS=linux go build github.com/go-delve/delve/cmd/dlv
```

Now we can copy it into the container like before:

```
$ docker cp ./dlv debugme-scratch:/dlv
```

However, we can't exec a shell because there is no shell inside the
container. We have to exec Delve directly:

```
$ docker exec -it debugme-scratch /dlv attach 1
Type 'help' for list of commands.
(dlv)
```

And away we Go!
