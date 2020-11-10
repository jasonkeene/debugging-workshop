
This program resolves the IP address of the hostname, typically `127.0.0.1`.
Unfortunately, it has a bug that causes it to panic. A core dump file was
captured for this crash. See if you can debug the issue from the core dump
alone. To get started run:

```
$ gunzip core.gz
$ dlv core binary core
```

<details>
  <summary>Hint 1</summary>

  Use `bt` to show a stack of where the panic happened. There is a lot of
  runtime code listed. It is usually a good idea to look for non-runtime code
  and find code that is specific to your program. Are there any frames here
  that look interesting?
</details>

<details>
  <summary>Hint 2</summary>

  Use `frame <m>` to set your frame to where the panic happened. This seems
  to be the correct frame:

  ```
  14  0x000000000053c706 in main.main
    at ./main.go:17
  ```

  Once the frame is set look around to see why this panic happened.
</details>

<details>
  <summary>Solution</summary>

  If you inspect the `err` value it was set to `nil` yet the panic still happened:

  ```
  (dlv) p err
  error(*main.codeError) nil
  ```

  Isn't that strange? As it turns out this value isn't a true `nil`, it is
  actually a typed `nil` with the type `*main.codeError`. This additional
  type information means that when you compare it to `nil` using `err != nil`
  it will actually be not `nil`.

  The solution here is to change the return type of `resolve` to be the
  `error` interface.
</details>
