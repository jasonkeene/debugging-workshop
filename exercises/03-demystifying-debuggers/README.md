
## Demystifying Debuggers

Most of the functionality of a debugger is implemented by the
[ptrace][ptrace] syscall. This syscall allows a "tracer" process (the
debugger) to control a "tracee" process (your Go program). You can intercept
the tracee's signals and syscalls, read and write to its registers and memory
and single-step through the program. Since the debugger can write to the
memory of the tracee, this allows it to set breakpoints that will suspend
execution of the tracee once they are hit.

Let's go over how this works. First, let's take a look at a Go program's
instructions as it runs:

<details>
  <summary>Show Animation</summary>

  <img src="https://github.com/jasonkeene/debugging-workshop/blob/master/exercises/03-demystifying-debuggers/ptrace1.gif">
</details>

On the left, you see multiple bytes of the process's memory. Each instruction
of the program can encode into multiple bytes. Here we are just looking at
two instructions. As the program is running the instruction pointer (the
white arrow) is advancing through memory as the CPU executes instructions.

To set a breakpoint, Delve finds the address in memory where the instruction
starts where we want to set a breakpoint. It then saves a copy of the first
byte of that instruction to use later. It then writes a trap instruction (a
software interrupt) on top of the original instruction. This is what that
looks like:

<details>
  <summary>Show Animation</summary>

  <img src="https://github.com/jasonkeene/debugging-workshop/blob/master/exercises/03-demystifying-debuggers/ptrace2.gif">
</details>

When this trap instruction is executed, the tracee is suspended and the
kernel's interrupt handler will get called, ultimately allowing the debugger
to take control:

<details>
  <summary>Show Animation</summary>

  <img src="https://github.com/jasonkeene/debugging-workshop/blob/master/exercises/03-demystifying-debuggers/ptrace3.gif">
</details>

Delve can now run any commands you give it. When you tell the debugger to
continue it:

1. Writes the original instruction back over the trap
1. Sets the tracee's program counter register to that memory address
1. Single steps the tracee to execute that instruction
1. Writes the trap back (so the breakpoint remains in place)
1. Resumes execution of the tracee

This is what that looks like:

<details>
  <summary>Show Animation</summary>

  <img src="https://github.com/jasonkeene/debugging-workshop/blob/master/exercises/03-demystifying-debuggers/ptrace4.gif">
</details>

So cool! And you thought debuggers were magic!

If you are interested in learning more about debuggers I found this series on
[Writing a Linux Debugger][writing-a-debugger] super interesting!

[ptrace]: http://man7.org/linux/man-pages/man2/ptrace.2.html
[writing-a-debugger]: https://blog.tartanllama.xyz/writing-a-linux-debugger-setup/
