
This program makes sequential requests to Google. The requests time out after
75ms. See if you can detect the goroutine leak and fix it.

<details>
  <summary>Solution</summary>
  
  The channel where the response is written to is unbuffered. The main
  goroutine will never read from this channel if a timeout occurs and so the
  goroutine will be blocked and never exit.

  You can increase the buffer size of the channel to prevent it from
  blocking. Alternatively, you can use a request context or http.Client
  timeout fields to timeout the request.
</details>
