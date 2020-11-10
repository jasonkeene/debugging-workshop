
This program is a concurrent ring buffer. It allows data to be read and
written to the buffer from multiple goroutines via a shared mutex. However,
for some reason, the order that we put data in is not the order that it comes
out in. This is very unexpected.

This debugme is part of the [Automating Delve][07] exercise. Feel free to
follow along there.

<details>
  <summary>Hint 1</summary>

  After running the Go client program we discovered that when the warning is
  occurring the delta between the indices is `20001`. `20001` is one greater
  than the capacity of the ring buffer.
</details>

<details>
  <summary>Hint 2</summary>

  When the delta is greater than the capacity of the buffer that means the
  writer has written so much data that it has filled up the buffer and has
  also overwritten values that have not been read.
</details>

[07]: https://github.com/jasonkeene/debugging-workshop/blob/master/exercises/07-automating-delve
