
This program uses a simple "framework" to schedule asynchronous work.

The framework waits for a signal start to cyclically execute a user given handler. Each time it is called, the provided handler may return a result or an error.

The framework supports graceful shutdown: when the process receives a SIGINT, the asynchronous worker is stopped and everything is cleaned up.
Before terminating, a goodbye message is written to standard output.

Unfortunately, the graceful shutdown seems not to work. Can you find out why?

<details>
  <summary>Hint</summary>

  What happen when the signal is received and the asynchronous worker randomize an error?
</details>

<details>
  <summary>Solution</summary>

  When the main goroutine receives the SIGINT signal, it exits from the `for` - `select` loop. Then, it signals the background worker to stop using `cancel()`, before waiting on the WaitGroup at line 111.
  But if this happens when the worker goroutine is sleeping, when returning from the handler, it will try to send back a result or an error. If it tries to send back an error, it will remain blocked on the channel send, since the `errs` channel is unbuffered.

  Substitute:

  ```go
  errs := make(chan error)
  ```

  with:

  ```go
  errs := make(chan error, 1)
  ```
</details>
