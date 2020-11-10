
This program manipulates an array of 10,000 integers concurrently. When an
element in the array needs to be accessed a lock is acquired specifically for
that element. This prevents data races and decreases contention for a global
lock. Unfortunately, after the program runs successfully for some time, it
will crash with a deadlock message:

```
$ go run .
2: [6074 6162 7671 4719 1839 8449 2982 6975 3681 8072]
5: [4890 9964 6940 9925 9183 8954 9226 9041 3605 9469]
3: [9376 8845 1329 9102 3296 5846 3227 263 2654 6266]
7: [4441 9465 8642 6744 1563 8034 5755 8196 1705 9771]
4: [2056 1974 807 3439 9586 7811 7595 8194 9290 1553]
0: [3601 8547 56 5449 6907 8752 9117 8080 8746 8348]
6: [6612 8325 9656 62 3193 6888 4470 1950 9234 5715]
1: [6511 7340 3028 8513 1474 6694 7159 3793 6545 7755]
fatal error: all goroutines are asleep - deadlock!
```

Figure out why this program is deadlocking and what the fix is.

<details>
  <summary>Hint 1</summary>

  Try running the program with the debugger attached:

  ```
  $ dlv debug
  Type 'help' for list of commands.
  (dlv) c
  ```

  When the deadlock panic occurs delve will prevent the program from crashing
  and allow you to debug the issue:

  ```
  fatal error: all goroutines are asleep - deadlock!
  > [runtime-fatal-throw] runtime.fatalthrow() /usr/local/go/src/runtime/panic.go:1162 (hits total:1) (PC: 0x10389a0)
  Warning: debugging optimized function
    1157: // fatalthrow implements an unrecoverable runtime throw. It freezes the
    1158: // system, prints stack traces starting from its caller, and terminates the
    1159: // process.
    1160: //
    1161: //go:nosplit
  =>1162: func fatalthrow() {
    1163:         pc := getcallerpc()
    1164:         sp := getcallersp()
    1165:         gp := getg()
    1166:         // Switch to the system stack to avoid any stack growth, which
    1167:         // may make things worse if the runtime is in a bad state.
  (dlv)
  ```

  Try poking around the stacks to see if you can find the cause.
</details>

<details>
  <summary>Hint 2</summary>

  Sometimes it is easier to figure out the cause of the bug by reproducing it
  with a smaller set of data.
</details>
