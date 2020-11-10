
## Introduction

Debugging is the process of understanding a program's execution for the
specific purpose of solving bugs. There are many tools and techniques we can
use to understand our programs better.

### Print Debugging

The most common debugging technique is to simply print information to
stdout/stderr. This is quite effective and has the advantage that it works in
almost all situations. Nearly every program can print. No matter how you
deploy your programs you will most likely be able to capture printed output
in some way.

This technique can quickly get out of hand, however. While iterating on a bug
locally you may find yourself littering `println()` calls throughout your
code and library code. Modifying source files that are marked as read-only
and are shared globally is non-ideal. Cleaning up these statements can be a
hassle and as you add more and more output it can be increasingly difficult
to understand what is going on.

Print debugging is typically the first tool I will reach for but will quickly
shift away from once I've hit four or five print lines or need to inspect the
state of library code.

### Debuggers

Debuggers are programs that allow you to manipulate the execution of other
programs. They allow you to:

- Set breakpoints that halt the program
- Step through the execution of the program
- Inspect and modify the state of the program
- Resume execution of the program

The debugger we'll be using in this workshop is Delve. It was created
specifically for Go and works great!

### Other Tools

Debuggers are not the only useful tool for debugging. In later examples we
will use pprof and bpftrace which are useful tools that fall under the realm
of profiling and tracing respectively.

Additionally, you might add static instrumentation to your code. In
production environments it is common to capture statistical metrics about
your programs and to link together spans from multiple systems to trace
request paths. These techniques are incredibly useful but will not be the
focus of this workshop.

### Run `dlv`

Make sure you have a recent version of Delve installed. The most recent
version as of writing is `1.5.0`:

```
$ dlv version
Delve Debugger
Version: 1.5.0
Build: $Id: ca5318932770ca063fc9885b4764c30bfaf8a199 $
```

Awesome, now open the `main.go` file in this directory in your editor so we
can follow along with what Delve is doing. Use Delve to debug the program in
this directory:

```
$ cd exercises/00-introduction
$ dlv debug
Type 'help' for list of commands.
(dlv)
```

Running `dlv debug` will compile the program and start it with the debugger
attached. You now have a `(dlv)` prompt where you can enter commands for the
debugger. The program is not executing yet. You can type `continue` or just
`c` to start it:

```
(dlv) c
this runs before the static breakpoint
> main.main() ./main.go:13 (PC: 0x106624a)
     8:         // This is an example of adding a static breakpoint in your program.
     9:         runtime.Breakpoint()
    10:
    11:         // The debugger will stop execution before this line and so you will not see
    12:         // this output immediately.
=>  13:         println("this runs after the static breakpoint")
    14:
    15:         println("a bunch")
    16:         println("of other")
    17:         println("lines")
    18:
```

You will notice you hit a breakpoint after the line `runtime.Breakpoint()` in
the source code. This is a static breakpoint that you can add to your
program. I use them in the examples but most of the time you will not want to
add breakpoints statically to your program. They are a hassle to remove from
your code after you are done debugging but also if you accidentally leave
them in your code you will get `SIGTRAP: trace trap` panics when you are not
debugging. It is much better to add them dynamically from within Delve with
the `break` or `b` command. Let's add one. First, we need to see more of the
source code to know where to set our breakpoint. We can tell Delve to display
more source lines with the `config` command:

```
(dlv) config -list
aliases                map[]
substitute-path        []
max-string-len         <not defined>
max-array-values       <not defined>
max-variable-recurse   <not defined>
disassemble-flavor     <not defined>
show-location-expr     false
source-list-line-color 34
source-list-line-count <not defined>
debug-info-directories [/usr/lib/debug/.build-id]
```

This shows all the parameters you can configure. Let's set
`source-list-line-count` to be something larger than the default of 5:

```
(dlv) config source-list-line-count 10
(dlv) ls
> main.main() ./main.go:13 (PC: 0x46788a)
     3: import "runtime"
     4:
     5: func main() {
     6:         println("this runs before the static breakpoint")
     7:
     8:         // This is an example of adding a static breakpoint in your program.
     9:         runtime.Breakpoint()
    10:
    11:         // The debugger will stop execution before this line and so you will not see
    12:         // this output immediately.
=>  13:         println("this runs after the static breakpoint")
    14:
    15:         println("a bunch")
    16:         println("of other")
    17:         println("lines")
    18:
    19:         // BREAK HERE
    20:         println("add a breakpoint on this line")
    21: }
(dlv)
```

We can now see we want to set our breakpoint on line 20. We can not set it on
19 because that is a comment.

```
(dlv) b 20
Breakpoint 1 set at 0x10662d6 for main.main() ./main.go:20
```

This adds a breakpoint just before line 20. There are many other ways of
telling Delve where to place a breakpoint. We will dive into those later. For
now let's run the program until we hit the breakpoint we just set:

```
(dlv) c
this runs after the static breakpoint
a bunch
of other
lines
> main.main() ./main.go:20 (hits goroutine(1):1 total:1) (PC: 0x10662d6)
    15:         println("a bunch")
    16:         println("of other")
    17:         println("lines")
    18:
    19:         // BREAK HERE
=>  20:         println("add a breakpoint on this line")
    21: }
```

Finally, you can continue to finish execution of the program:

```
(dlv) c
add a breakpoint on this line
Process 78900 has exited with status 0
```

To exit Delve type `quit` or `q`:

```
(dlv) q
```
