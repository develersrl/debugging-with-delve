
## Taming Concurrency

Go has a quite unique model for concurrency. It is based on:
- goroutines: user-space lightweigth threads managed by the Go runtime
- channels: typed conduit through which goroutines can send and receive values

Delve offers a great support for Go concurrency model, as it is able to follow
the execution flow between the various goroutines, show them at each time,
and separating the ones started by the user from the ones started by
the runtime itself (e.g. the ones running the garbage collector).

Specifically, Delve is able to cope with two unique Go runtime features that
makes the debugging with [gdb](https://www.gnu.org/software/gdb/)
or [lldb](https://lldb.llvm.org/) less intuitive.
These features are the **work-stealing scheduler** and the goroutines
**contiguous stack**.
We'll see more about this features in the [Debuggers Under The Hood][01]
section.

Time to start another debugging session!

Open the file [main.go][02] in your editor to get familiar with the code.
As you can see, it is quite simple: the main goroutine acts as a consumer,
accumulating all the vaues received through the channel. The other goroutine
acts as a producer, randomizing values until 42 appears and the goroutine
returns.

Now, let's start Delve and, as usual, continue the execution until the `main`
function.

```
$ dlv debug
Type 'help' for list of commands.
(dlv) b main.main
Breakpoint 1 set at 0x49952f for main.main() ./main.go:13
(dlv) c
> main.main() ./main.go:13 (hits goroutine(1):1 total:1) (PC: 0x49952f)
     8:	
     9:	func init() {
    10:		rand.Seed(time.Now().Unix())
    11:	}
    12:	
=>  13:	func main() {
    14:		ch := make(chan int)
    15:	
    16:		go func() {
    17:			defer close(ch)
    18:	
(dlv) 
```

Delve has a command to show all the started threads. You'll find this familiar
if you come from the gdb world.

```
(dlv) threads
* Thread 8220 at 0x49952f ./main.go:13 main.main
  Thread 8224 at 0x462ebd /usr/local/go/src/runtime/sys_linux_amd64.s:149 runtime.usleep
  Thread 8225 at 0x4634a3 /usr/local/go/src/runtime/sys_linux_amd64.s:553 runtime.futex
  Thread 8226 at 0x4634a3 /usr/local/go/src/runtime/sys_linux_amd64.s:553 runtime.futex
(dlv)
```

This information is not that useful in our case: we want to reason about goroutines.
This is where Delve shines.

```
(dlv) goroutines
* Goroutine 1 - User: ./main.go:13 main.main (0x49952f) (thread 8220)
  Goroutine 2 - User: /usr/local/go/src/runtime/proc.go:362 runtime.gopark (0x4378d2) [force gc (idle)]
  Goroutine 3 - User: /usr/local/go/src/runtime/proc.go:362 runtime.gopark (0x4378d2) [GC sweep wait]
  Goroutine 4 - User: /usr/local/go/src/runtime/proc.go:362 runtime.gopark (0x4378d2) [GC scavenge wait]
  Goroutine 5 - User: /usr/local/go/src/runtime/proc.go:362 runtime.gopark (0x4378d2) [finalizer wait]
[5 goroutines]
(dlv) 
```

The `*` tells us that we are on goroutine 1, where we are stuck at `main.main`.
In fact, this is the only goroutine executing user code:

```
(dlv) goroutines -with user
* Goroutine 1 - User: ./main.go:13 main.main (0x49952f) (thread 8220)
[1 goroutines]
(dlv) 
```

The other ones belong to the Go runtime:

```
(dlv) goroutines -without user
  Goroutine 2 - User: /usr/local/go/src/runtime/proc.go:362 runtime.gopark (0x4378d2) [force gc (idle)]
  Goroutine 3 - User: /usr/local/go/src/runtime/proc.go:362 runtime.gopark (0x4378d2) [GC sweep wait]
  Goroutine 4 - User: /usr/local/go/src/runtime/proc.go:362 runtime.gopark (0x4378d2) [GC scavenge wait]
  Goroutine 5 - User: /usr/local/go/src/runtime/proc.go:362 runtime.gopark (0x4378d2) [finalizer wait]
[4 goroutines]
(dlv) 
```

This ability to separate user and runtime code is very handy: almost always the
issue is in your code, not in the runtime code. Being able to separate them
effectively, greatly reduce the noise while debugging.

Now, let's put a breakpoint on line 32, where the producer accumulates values.
Continue execution to hit it and inspect the goroutines once more.

```
(dlv) b main.go:32
Breakpoint 2 set at 0x4995e7 for main.main() ./main.go:32
(dlv) c
> main.main() ./main.go:32 (hits goroutine(1):1 total:1) (PC: 0x4995e7)
    27:			}
    28:		}()
    29:	
    30:		sum := 0
    31:		for v := range ch {
=>  32:			sum += v
    33:		}
    34:	
    35:		fmt.Printf("sum is %d\n", sum)
    36:	}
(dlv) goroutines -with user
* Goroutine 1 - User: ./main.go:32 main.main (0x4995e7) (thread 8220)
  Goroutine 6 - User: ./main.go:26 main.main.func1 (0x49979b) [chan send]
[2 goroutines]
```

As expected, we see two user goroutines.
The goroutine 1 is the one that hit the breakpoint at line 32. It is currently
running on an OS thread, as Delve is reporting to us with `(thread 3089)`.
The other one, just every each other goroutine, has been stopped as well.
We can see that goroutine 6 is not currently executing. Instead it has been
blocked while sending the next randomized value through the channel. This is
reported by Delve with `[chan send]` at the end of the line.

Let's confirm this analyzing the stack traces of both goroutines.
For goroutine 1:

```
(dlv) bt
0  0x00000000004995e7 in main.main
   at ./main.go:32
1  0x00000000004374b8 in runtime.main
   at /usr/local/go/src/runtime/proc.go:250
2  0x0000000000461601 in runtime.goexit
   at /usr/local/go/src/runtime/asm_amd64.s:1571
(dlv) 
```

For goroutine 6:

```
(dlv) goroutine 6 bt
0  0x00000000004378d2 in runtime.gopark
   at /usr/local/go/src/runtime/proc.go:362
1  0x0000000000404e1c in runtime.chansend
   at /usr/local/go/src/runtime/chan.go:258
2  0x0000000000404b9d in runtime.chansend1
   at /usr/local/go/src/runtime/chan.go:144
3  0x000000000049979b in main.main.func1
   at ./main.go:26
4  0x0000000000461601 in runtime.goexit
   at /usr/local/go/src/runtime/asm_amd64.s:1571
(dlv) 
```

Goroutine 6 has been parked by the runtime while sending the value.
This is happening because the `ch` channel is unbuffered, so the sending
goroutine and the receiving one must "meet" to pass the value between
themselves. In concurrent programming, this is called a **rendezvous barrier**.

The value already received from goroutine 1 is easy to see:

```
(dlv) print v
53
```

But what about the value that goroutine 6 is sending through the channel,
waiting for goroutine 1 to receive it?
We can prepend the `print` command with a specification of which goroutine and
which frame in the call stack we want to execute the command for.
We already did it before, for the `bt` command. We saw that the frame 3 was the
one representing `main.main.func1`, that is, the function associated with
goroutine 6, started by our statement `go func() { ... }()` at line
`main.go:16`.
So, to print the next value that we'll receive in the main goroutine:

```
(dlv) goroutine 6 frame 3 print v
44
```

That's really useful to debug concurrent programs: you can inspect frames from
whichever goroutine you want, without jump between them at every command.

Let's confirm what we observed. Since we still have a breakpoint on line 32, if
we hit `continue`, we'll break at the next `for` loop iteration, where we will
have received the next value from the producer goroutine.

```
(dlv) c
> main.main() ./main.go:32 (hits goroutine(1):2 total:2) (PC: 0x4995e7)
    27:			}
    28:		}()
    29:	
    30:		sum := 0
    31:		for v := range ch {
=>  32:			sum += v
    33:		}
    34:	
    35:		fmt.Printf("sum is %d\n", sum)
    36:	}
(dlv) print v
44
```

Very well, we were right!

Now that you learned how to tame concurrent programs with Delve, let's try
to put this new knowledge at work. Time for the second exercise!

[01]: https://github.com/develersrl/debugging-with-delve/blob/main/01-debugging/07-debuggers-under-the-hood
[02]: https://github.com/develersrl/debugging-with-delve/blob/main/01-debugging/02-taming-concurrency/main.go