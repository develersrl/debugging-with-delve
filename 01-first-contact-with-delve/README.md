
## First Contact With Delve

> Let's work the problem people. Let's not make things worse by guessing.
> - Gene Kranz

### Introduction

The most common debugging technique is to simply print information to
stdout/stderr, using `fmt.Println` and friends.
To be honest, this simple approach is often effective and has the additional
benefit of working in almost every environment. Besides, it does not impose
any additional burden on the developer: you have no additional tool to learn
to add some `fmt.Println` in your code.

The speed of the Go compiler and the ability to easily build and deploy for
multiple architectures are of great support while using this technique.

Besides, the Go tooling makes writing tests straightforward: well tested
codebases show less bugs than untested ones. This, in turn, reduces (but does
not eliminate) the need for a debugger.

Yet, sooner or later, you will find hard and unexpected bugs. Knowing another
tool, specifically a powerful one like Delve, can make a great difference when
you need to debug a production outage.

Delve has specific support for the Go programming language, and has everything
you need to tackle bugs in a more structured and systematic approach.

This workshop introduces the use of the Delve debugger so that when you will
face your next bug, you will have another powerful tool under you belt.

### Run `dlv`

Make sure you have a recent version of Delve installed. The most recent
version as of writing is `1.8.3`:

```
$ dlv version
Delve Debugger
Version: 1.8.3
Build: $Id: f92bb46b82b3b92d79ce59c4b55eeefbdd8d040c $
```

Awesome! Now, use Delve to debug the program in this directory:

```
$ cd 01-first-contact-with-delve
$ dlv debug
Type 'help' for list of commands.
(dlv) 
```

Running `dlv debug` will compile the program and start it with the debugger
attached. You now have a `(dlv)` prompt where you can enter commands for the
debugger. The program is not executing yet.

The message just before the `(dlv)` prompt gives us a nice information: if we
are unsure about what to do, typing `help` may be of great... well, help.

```
$ dlv debug
Type 'help' for list of commands.
(dlv) help
The following commands are available:

Running the program:
    call ------------------------ Resumes process, injecting a function call (EXPERIMENTAL!!!)
    continue (alias: c) --------- Run until breakpoint or program termination.
    next (alias: n) ------------- Step over to next source line.
    rebuild --------------------- Rebuild the target executable and restarts it. It does not work if the executable was not built by delve.
    restart (alias: r) ---------- Restart process.
    step (alias: s) ------------- Single step through program.
    step-instruction (alias: si)  Single step a single cpu instruction.
    stepout (alias: so) --------- Step out of the current function.

Manipulating breakpoints:
    break (alias: b) ------- Sets a breakpoint.
    breakpoints (alias: bp)  Print out info for active breakpoints.
    clear ------------------ Deletes breakpoint.
    clearall --------------- Deletes multiple breakpoints.
    condition (alias: cond)  Set breakpoint condition.
    on --------------------- Executes a command when a breakpoint is hit.
    toggle ----------------- Toggles on or off a breakpoint.
    trace (alias: t) ------- Set tracepoint.
    watch ------------------ Set watchpoint.

Viewing program variables and memory:
    args ----------------- Print function arguments.
    display -------------- Print value of an expression every time the program stops.
    examinemem (alias: x)  Examine raw memory at the given address.
    locals --------------- Print local variables.
    print (alias: p) ----- Evaluate an expression.
    regs ----------------- Print contents of CPU registers.
    set ------------------ Changes the value of a variable.
    vars ----------------- Print package variables.
    whatis --------------- Prints type of an expression.

Listing and switching between threads and goroutines:
    goroutine (alias: gr) -- Shows or changes current goroutine
    goroutines (alias: grs)  List program goroutines.
    thread (alias: tr) ----- Switch to the specified thread.
    threads ---------------- Print out info for every traced thread.

Viewing the call stack and selecting frames:
    deferred --------- Executes command in the context of a deferred call.
    down ------------- Move the current frame down.
    frame ------------ Set the current frame, or execute command on a different frame.
    stack (alias: bt)  Print stack trace.
    up --------------- Move the current frame up.

Other commands:
    config --------------------- Changes configuration parameters.
    disassemble (alias: disass)  Disassembler.
    dump ----------------------- Creates a core dump from the current process state
    edit (alias: ed) ----------- Open where you are in $DELVE_EDITOR or $EDITOR
    exit (alias: quit | q) ----- Exit the debugger.
    funcs ---------------------- Print list of functions.
    help (alias: h) ------------ Prints the help message.
    libraries ------------------ List loaded dynamic libraries
    list (alias: ls | l) ------- Show source code.
    source --------------------- Executes a file containing a list of delve commands
    sources -------------------- Print list of source files.
    transcript ----------------- Appends command output to a file.
    types ---------------------- Print list of types

Type help followed by a command for full documentation.
```

Quite a lot of commands. We will walk around most of them during the workshop.

Before starting to debug, let's take a look at the code. Every Go program must
have a `main` function in the package `main`. Let's start from there.

```
(dlv) list main.main
Showing /home/vagrant/debugging-with-delve/01-first-contact-with-delve/main.go:25 (PC: 0x4998f2)
    20:	
    21:	func sum(a, b int) int {
    22:		return a + b
    23:	}
    24:	
    25:	func main() {
    26:		a, b := 10, 5
    27:		total := sum(a, b)
    28:		fmt.Printf("%d + %d = %d\n", a, b, total)
    29:	
    30:		p := Progression{
```

The `list` command (or `ls`, if you don't like to type) showed us an excerpt of
the `main` function.
That's nice, but I'd like to see more lines. How can we tell Delve to show more?
Broadly speaking, how can we configure Delve?

```
(dlv) config -list
aliases                   map[]
substitute-path           []
max-string-len            <not defined>
max-array-values          <not defined>
max-variable-recurse      <not defined>
disassemble-flavor        <not defined>
show-location-expr        false
source-list-line-color    <nil>
source-list-arrow-color   ""
source-list-keyword-color ""
source-list-string-color  ""
source-list-number-color  ""
source-list-comment-color ""
source-list-line-count    <not defined>
debug-info-directories    [/usr/lib/debug/.build-id]
```

The option `source-list-line-count` seems to be the one we are searching for!
Let's try to setup that to 50.

```
(dlv) config source-list-line-count 50
```

And now list again all the config options.

```
(dlv) config -list
aliases                   map[]
substitute-path           []
max-string-len            <not defined>
max-array-values          <not defined>
max-variable-recurse      <not defined>
disassemble-flavor        <not defined>
show-location-expr        false
source-list-line-color    <nil>
source-list-arrow-color   ""
source-list-keyword-color ""
source-list-string-color  ""
source-list-number-color  ""
source-list-comment-color ""
source-list-line-count    50
debug-info-directories    [/usr/lib/debug/.build-id]
```

We did it! But, will this be permanent?
Let's exit from Delve giving it a *Ctrl-D* and retry again to list the
`main.main` symbol.

```
(dlv) list main.main
Showing /home/vagrant/debugging-with-delve/01-first-contact-with-delve/main.go:25 (PC: 0x4998f2)
    20:	
    21:	func sum(a, b int) int {
    22:		return a + b
    23:	}
    24:	
    25:	func main() {
    26:		a, b := 10, 5
    27:		total := sum(a, b)
    28:		fmt.Printf("%d + %d = %d\n", a, b, total)
    29:	
    30:		p := Progression{
```

We are at it again. Listing the config options confirms that.

```
(dlv) config -list

...

source-list-line-count    <not defined>

...

```

Why? That's because Delve starts reading a configuration file from a specific location.

From the Delve official documentation:

> If `$XDG_CONFIG_HOME` is set, then configuration and command history files are located in `$XDG_CONFIG_HOME/dlv`. Otherwise, they are located in `$HOME/.config/dlv` on Linux and `$HOME/.dlv` on other systems.
> 
> The configuration file `config.yml` contains all the configurable options and their default values. The command history is stored in `.dbg_history`.

Now we can see the starting value of `source-list-line-count`.

```
$ cat $HOME/.config/dlv/config.yml
# Configuration file for the delve debugger.

# This is the default configuration file. Available options are provided, but disabled.
# Delete the leading hash mark to enable an item.

...

# Uncomment to change the number of lines printed above and below cursor when
# listing source code.
# source-list-line-count: 5

...

```

So, if you'd like to change it permanently, change it here. Otherwise, you will
have to change the value at each debugging session. That's not bad, since you
will always have your commands history.
With the UP key you can scroll your past commands and with *Ctrl-R* you can use
the reverse search, just like in your shell.

Now, back to where we left our debugging session.

As we've seen, after entering `dlv debug`, we see the Delve prompt, but the
program is not running yet.
Let's run the program to have a sense of what it will ll do:

```
(dlv) c
10 + 5 = 15
Process 8131 has exited with status 0
```

Fine. Now we'll restart it and finally begin to unleash the debugging power of Delve.

```
(dlv) restart
Process restarted with PID 8139
```

First, let's put a breakpoint just at the `main` function:

```
(dlv) b main.main
Breakpoint 1 set at 0x4998f2 for main.main() ./main.go:25
```

We can list all breakpoints currently set with:

```
(dlv) breakpoints
Breakpoint runtime-fatal-throw (enabled) at 0x435060 for runtime.throw() /usr/local/go/src/runtime/panic.go:982 (0)
Breakpoint unrecovered-panic (enabled) at 0x435420 for runtime.fatalpanic() /usr/local/go/src/runtime/panic.go:1065 (0)
	print runtime.curg._panic.arg
Breakpoint 1 (enabled) at 0x4998f2 for main.main() ./main.go:25 (0)
```

The first two are set automatically by Delve, the one with ID 1 is ours.

To resume the program execution:

```
(dlv) c
> main.main() ./main.go:25 (hits goroutine(1):1 total:1) (PC: 0x497d72)
    20:	
    21:	func sum(a, b int) int {
    22:		return a + b
    23:	}
    24:	
=>  25:	func main() {
    26:		a, b := 10, 5
    27:		total := sum(a, b)
    28:		fmt.Printf("%d + %d = %d\n", a, b, total)
    29:	
    30:		p := Progression{
```

As expected, we break at `main.main` and Delve kindly shows us the listing.
Notice the arrow, suggesting you the current position of the Instruction Pointer
register.

Let's step over every single instruction. We can do this using `next` or `n`.

```
(dlv) n
> main.main() ./main.go:26 (PC: 0x499909)
    21:	func sum(a, b int) int {
    22:		return a + b
    23:	}
    24:	
    25:	func main() {
=>  26:		a, b := 10, 5
    27:		total := sum(a, b)
    28:		fmt.Printf("%d + %d = %d\n", a, b, total)
    29:	
    30:		p := Progression{
    31:			Value: 1,
```

The arrow advanced as expected. But what if we want to step into the `sum`
function? To do that, once we are on the 27th line, we can use `step` or `s`
to follow the call to `sum`.

```
> main.sum() ./main.go:21 (PC: 0x4998a0)
    16:	
    17:	func (p *Progression) Next() {
    18:		p.Value += p.Step
    19:	}
    20:	
=>  21:	func sum(a, b int) int {
    22:		return a + b
    23:	}
    24:	
    25:	func main() {
    26:		a, b := 10, 5
```

It would be nice to get the values of all the function arguments.

```
(dlv) args
a = 10
b = 5
~r0 = 518
```

As we've seen in the source code for `main.main`, `sum` has been called with
a = 10 and b = 5.
To step out of this function and return back to `main.main`, use `stepout` or
`so`.

Is there a way to print a specific variable? Sure:

```
(dlv) print a
10
(dlv) print b
5
```

What we've used up until now, are features that belongs to all full-fledged
debuggers. Why should we use Delve for our Go programs, then?
Let's see how Delve is capable of understanding interfaces, slices, maps and
channels!

Step or continue until line 41 and see how Delve prints the value of `c`.

```
(dlv) print c
main.Counter(*main.Progression) *{Value: 11, Step: 1}
```

Delve is telling us that the `c` variable:
- is a `main.Counter` interface
- it contains a pointer to a `main.Progression` type
- the struct pointed to holds the values `{Value: 11, Step: 1}`

Time to see how Delve manages slices.

```
(dlv) b main.go:48
Breakpoint 1 set at 0x498026 for main.main() ./main.go:48
(dlv) c
> main.main() ./main.go:48 (hits goroutine(1):1 total:1) (PC: 0x498026)
    43:		// with slices...
    44:		s := make([]int, 10)
    45:		for i := 10; i > 0; i-- {
    46:			s[10-i] = i
    47:		}
=>  48:		sort.Ints(s)
    49:	
    50:		// with maps...
    51:		m := make(map[int]string)
    52:		for i := 0; i < 10; i++ {
    53:			m[i] = fmt.Sprintf("%d", i)
(dlv) print s
[]int len: 10, cap: 10, [10,9,8,7,6,5,4,3,2,1]
```

And, after the sorting:

```
(dlv) print s
[]int len: 10, cap: 10, [1,2,3,4,5,6,7,8,9,10]
```

Delve is able to manage each element of the slice.

```
(dlv) print s[5]
6
```

And, using `call` we can even inject a function call (there are limitations,
though).

```
(dlv) call s[5]=100
> main.main() ./main.go:51 (PC: 0x498045)
    46:			s[10-i] = i
    47:		}
    48:		sort.Ints(s)
    49:	
    50:		// with maps...
=>  51:		m := make(map[int]string)
    52:		for i := 0; i < 10; i++ {
    53:			m[i] = fmt.Sprintf("%d", i)
    54:		}
    55:		m[3] = "Debugging Workshop"
    56:	
(dlv) print s[5]
100
```

Map are handled well, too:

```
(dlv) print m
map[int]string [
	0: "0", 
	4: "4", 
	6: "6", 
	9: "9", 
	1: "1", 
	2: "2", 
	3: "Debugging Workshop", 
	5: "5", 
	7: "7", 
	8: "8", 
]
(dlv) print m[3]
"Debugging Workshop"
```

That's it for now. Let's close Delve with a *Ctrl-D*. Time for the first exercise!