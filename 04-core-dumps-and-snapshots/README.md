
## Core Dumps and Snapshots

A core dump is a serialized form of the current execution state of the
process. The name "core dump" originates from "core memory" which was a type
of magnetic memory used in early computers before silicon memory was created.

These dumps include the of process's memory, registers, flags, and some
operating system specific state. We can load these files into Delve and
inspect them.

⚠️ Since these dumps contain all the process's memory it is critical to handle
them securely. If there is confidential information in memory at the time the
core dump is taken, that will be part of the core dump.

There are two situations where you might want to debug from a core file:

- Post-mortem debugging
- Snapshot debugging

### Post-mortem Debugging

Let's start building the code in this folder, running it and querying the server.

```
$ go build -gcflags=all='-N -l' -o server
$ ./server 
2021/10/03 13:25:02 listening on localhost:12345
```

On another terminal:

```
$ curl localhost:12345
hello from 17179
```

The server is answering returning its pid. But if we query it using this URL
path:

```
$ curl localhost:12345/crash
curl: (52) Empty reply from server
```

The server has crashed:

```
$ ./server
2022/05/24 10:19:30 listening on localhost:12345
panic: :(

goroutine 23 [running]:
main.handler.func1()
	/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/main.go:21 +0x27
created by main.handler
	/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/main.go:20 +0xdf
```

When this happen, the Go runtime, by default, prints a stack trace for the
current goroutine and then exits with exit code 2. In addition, we can instruct
the runtime to generate a core dump.

First of all, we should check the current user limit for core dumps.

```
$ ulimit -c
0
```

No space on disk is available for core dumps. We can remove this limit with:

```
$ ulimit -c unlimited
$ ulimit -c
unlimited
```

Note that, doing this, we are updating the limit only for the current shell
session. To set it permanently, you should edit the appropriate line in
`/etc/security/limits.conf`.
For production servers, you want to avoid the `unlimited` value: core dumps may
be heavy and they can soon fill up your storage.

To let the Go runtime generate a core dump we must set the env variable
`GOTRACEBACK` to `crash`. This way, in case of failure, stack traces from all
goroutines will be printed and a core dump will be generated.

```
$ GOTRACEBACK=crash ./server
2022/05/24 10:20:51 listening on localhost:12345
panic: :(

...

goroutine 9 [runnable]:
net/http.(*connReader).startBackgroundRead.func2()
	/usr/local/go/src/net/http/server.go:668 fp=0xc0000387e0 sp=0xc0000387d8 pc=0x70b2c0
runtime.goexit()
	/usr/local/go/src/runtime/asm_amd64.s:1571 +0x1 fp=0xc0000387e8 sp=0xc0000387e0 pc=0x46a581
created by net/http.(*connReader).startBackgroundRead
	/usr/local/go/src/net/http/server.go:668 +0x179
Aborted (core dumped)
```

The core dumps are managed by systemd, so we are going to use the `coredumpctl`
utility to get them.

List all the available core dumps:

```
$ sudo coredumpctl list
TIME                         PID  UID  GID SIG     COREFILE EXE                                                                     SIZE
Tue 2022-05-24 10:20:56 UTC 9182 1000 1000 SIGABRT present  /home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server 166.9K
```

In this case, there is just the one we generated with the server.

The core dump is stored in compressed form. To extract it and save it locally:

```
$ sudo coredumpctl dump --output ./core
           PID: 9182 (server)
           UID: 1000 (vagrant)
           GID: 1000 (vagrant)
        Signal: 6 (ABRT)
     Timestamp: Tue 2022-05-24 10:20:55 UTC (2min 34s ago)
  Command Line: ./server
    Executable: /home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server
 Control Group: /user.slice/user-1000.slice/session-4.scope
          Unit: session-4.scope
         Slice: user-1000.slice
       Session: 4
     Owner UID: 1000 (vagrant)
       Boot ID: 0399c18323a544199fd1c826c5cba927
    Machine ID: 1ed5e33c14094d8da23d5e468bddf473
      Hostname: ubuntu-jammy
       Storage: /var/lib/systemd/coredump/core.server.1000.0399c18323a544199fd1c826c5cba927.9182.1653387655000000.zst (present)
     Disk Size: 166.9K
       Message: Process 9182 (server) of user 1000 dumped core.
                
                Found module /home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server without build-id.
                Found module linux-vdso.so.1 with build-id: 14d4ad7fcca497e197971da1f799d6fe0289272e
                Found module ld-linux-x86-64.so.2 with build-id: aa1b0b998999c397062e1016f0c95dc0e8820117
                Found module libc.so.6 with build-id: 89c3cb85f9e55046776471fed05ec441581d1969
                Stack trace of thread 9182:
                #0  0x000000000046bec1 n/a (/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server + 0x6bec1)
                #1  0x0000000000450aa5 n/a (/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server + 0x50aa5)
                #2  0x000000000044f4c7 n/a (/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server + 0x4f4c7)
                #3  0x000000000046cc6e n/a (/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server + 0x6cc6e)
                #4  0x000000000046c19d n/a (/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server + 0x6c19d)
                #5  0x00007fa50cb34520 __restore_rt (libc.so.6 + 0x42520)
                #6  0x000000000046bec1 n/a (/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server + 0x6bec1)
                #7  0x00000000004506b8 n/a (/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server + 0x506b8)
                #8  0x000000000043abe5 n/a (/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server + 0x3abe5)
                #9  0x000000000043a3ba n/a (/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server + 0x3a3ba)
                #10 0x000000000072d0e7 n/a (/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server + 0x32d0e7)
                #11 0x000000000046a581 n/a (/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/server + 0x6a581)
```

Here is our core dump:

```
$ ls -lh core
-rw-r--r-- 1 vagrant vagrant 82M May 24 10:23 core
```

As anticipated, it is quite heavy: 82 MB just for this small program.

Now, to start debugging from a core dump:

```
$ dlv core server core
Type 'help' for list of commands.
(dlv) bt
 0  0x000000000046bec1 in runtime.raise
    at /usr/local/go/src/runtime/sys_linux_amd64.s:168
 1  0x00000000004504a5 in runtime.dieFromSignal
    at /usr/local/go/src/runtime/signal_unix.go:852
 2  0x0000000000450aa5 in runtime.sigfwdgo
    at /usr/local/go/src/runtime/signal_unix.go:1066
 3  0x000000000044f4c7 in runtime.sigtrampgo
    at /usr/local/go/src/runtime/signal_unix.go:430
 4  0x000000000046cc6e in runtime.sigtrampgo
    at <autogenerated>:1
 5  0x000000000046c19d in runtime.sigtramp
    at /usr/local/go/src/runtime/sys_linux_amd64.s:361
 6  0x00007fa50cb34520 in ???
    at ?:-1
 7  0x00000000004506b8 in runtime.crash
    at /usr/local/go/src/runtime/signal_unix.go:944
 8  0x000000000043abe5 in runtime.fatalpanic
    at /usr/local/go/src/runtime/panic.go:1092
 9  0x000000000043a3ba in runtime.gopanic
    at /usr/local/go/src/runtime/panic.go:941
10  0x000000000072d0e7 in main.handler.func1
    at ./main.go:21
11  0x000000000046a581 in runtime.goexit
    at /usr/local/go/src/runtime/asm_amd64.s:1571
(dlv) 
```

In this kind of debugging session we can use the usual commands Delve offers
us, but remember that the core dump is a snapshot. You can't `continue` or
`next`, since the program is not still executing (it already crashed).
But you can inspect memory, registers, stack traces and so on.

This is not the only way to generate a core dump. We can induce a crash sending
a SIGQUIT to our process.
Restart the server:

```
$ GOTRACEBACK=crash ./server
2022/05/24 10:28:03 listening on localhost:12345
```

and, on another terminal:

```
$ kill -s SIGQUIT $(pidof server)
```

The server will crash and produce a core dump:

```
SIGQUIT: quit
PC=0x46c660 m=0 sigcode=0

...

-----

SIGQUIT: quit
PC=0x46c441 m=6 sigcode=0

goroutine 0 [idle]:
runtime.futex()
	/usr/local/go/src/runtime/sys_linux_amd64.s:552 +0x21 fp=0x7fdd17ffece8 sp=0x7fdd17ffece0 pc=0x46c441
runtime.futexsleep(0x40e299?, 0x950540?, 0x7fdd17ffed70?)
	/usr/local/go/src/runtime/os_linux.go:66 +0x36 fp=0x7fdd17ffed48 sp=0x7fdd17ffece8 pc=0x4368d6
runtime.notesleep(0x950558)
	/usr/local/go/src/runtime/lock_futex.go:159 +0x9d fp=0x7fdd17ffed80 sp=0x7fdd17ffed48 pc=0x40ddbd
runtime.templateThread()
	/usr/local/go/src/runtime/proc.go:2206 +0x69 fp=0x7fdd17ffeda8 sp=0x7fdd17ffed80 pc=0x440c49
runtime.mstart1()
	/usr/local/go/src/runtime/proc.go:1418 +0x93 fp=0x7fdd17ffedd0 sp=0x7fdd17ffeda8 pc=0x43f973
runtime.mstart0()
	/usr/local/go/src/runtime/proc.go:1376 +0x7e fp=0x7fdd17ffee00 sp=0x7fdd17ffedd0 pc=0x43f89e
runtime.mstart()
	/usr/local/go/src/runtime/asm_amd64.s:367 +0x5 fp=0x7fdd17ffee08 sp=0x7fdd17ffee00 pc=0x4682c5
rax    0xca
rbx    0x0
rcx    0x46c443
rdx    0x0
rdi    0x950558
rsi    0x80
rbp    0x7fdd17ffed38
rsp    0x7fdd17ffece0
r8     0x0
r9     0x0
r10    0x0
r11    0x286
r12    0xc000003a00
r13    0x16
r14    0xc000003a00
r15    0x7ffe706dc880
rip    0x46c441
rflags 0x286
cs     0x33
fs     0x0
gs     0x0
Aborted (core dumped)
```

### Snapshot Debugging

It is possible to get a core dump even without crashing the process. These are
called snapshots. This may be handy if we don't want to keep a debugger
attached to the process, to avoid slowing down its execution too much.

We can generate a snapshot using `gcore`.

```
$ gcore $(pidof server)
[New LWP 9555]
[New LWP 9556]
[New LWP 9557]
[New LWP 9558]

...

Saved corefile core.9554
[Inferior 1 (process 9554) detached]
```

Now we can start debugging from the core dump as seen before. Note that the
process of generating a core dump can take several seconds, but once you
generated it, you will be able to debug without interrupting your process
anymore.

### Reproducible builds

When building executables for production deploy, you usually want to keep the
optimizations enabled, strip your binaries and remove all the file system
paths.
We've already seen how the first choice may impact your debugging sessions and
we've also seen how to overcome the issues that may arise (at least to a
certain extent).
Let's investigated more how a stripped binary may impact your debugging.

First of all, what does it mean to strip a binary?
Let's build an executable the usual way and inspect it with `readelf`:

```
$ go build -gcflags=all='-N -l' -o server
$ readelf -S server 
There are 36 section headers, starting at offset 0x270:

Section Headers:
  [Nr] Name              Type             Address           Offset
       Size              EntSize          Flags  Link  Info  Align
  [ 0]                   NULL             0000000000000000  00000000
       0000000000000000  0000000000000000           0     0     0
  [ 1] .text             PROGBITS         0000000000401000  00001000
       000000000032c35f  0000000000000000  AX       0     0     32
  [ 2] .plt              PROGBITS         000000000072d360  0032d360
       0000000000000220  0000000000000010  AX       0     0     16
  [ 3] .rodata           PROGBITS         000000000072e000  0032e000
       00000000000a5958  0000000000000000   A       0     0     32
  [ 4] .rela             RELA             00000000007d3958  003d3958
       0000000000000018  0000000000000018   A      11     0     8
  [ 5] .rela.plt         RELA             00000000007d3970  003d3970
       0000000000000318  0000000000000018   A      11     2     8
  [ 6] .gnu.version      VERSYM           00000000007d3ca0  003d3ca0
       000000000000004c  0000000000000002   A      11     0     2
  [ 7] .gnu.version_r    VERNEED          00000000007d3d00  003d3d00
       0000000000000070  0000000000000000   A      10     1     8
  [ 8] .hash             HASH             00000000007d3d80  003d3d80
       00000000000000bc  0000000000000004   A      11     0     8
  [ 9] .shstrtab         STRTAB           0000000000000000  003d3e40
       00000000000001e6  0000000000000000           0     0     1
  [10] .dynstr           STRTAB           00000000007d4040  003d4040
       0000000000000230  0000000000000000   A       0     0     1
  [11] .dynsym           DYNSYM           00000000007d4280  003d4280
       0000000000000390  0000000000000018   A      10     1     8
  [12] .typelink         PROGBITS         00000000007d4620  003d4620
       0000000000001284  0000000000000000   A       0     0     32
  [13] .itablink         PROGBITS         00000000007d58c0  003d58c0
       00000000000007b8  0000000000000000   A       0     0     32
  [14] .gosymtab         PROGBITS         00000000007d6078  003d6078
       0000000000000000  0000000000000000   A       0     0     1
  [15] .gopclntab        PROGBITS         00000000007d6080  003d6080
       000000000010afb8  0000000000000000   A       0     0     32
  [16] .go.buildinfo     PROGBITS         00000000008e2000  004e2000
       0000000000000220  0000000000000000  WA       0     0     16
  [17] .dynamic          DYNAMIC          00000000008e2220  004e2220
       0000000000000120  0000000000000010  WA      10     0     8
  [18] .got.plt          PROGBITS         00000000008e2340  004e2340
       0000000000000120  0000000000000008  WA       0     0     8
  [19] .got              PROGBITS         00000000008e2460  004e2460
       0000000000000008  0000000000000008  WA       0     0     8
  [20] .noptrdata        PROGBITS         00000000008e2480  004e2480
       0000000000030e78  0000000000000000  WA       0     0     32
  [21] .data             PROGBITS         0000000000913300  00513300
       000000000000bc10  0000000000000000  WA       0     0     32
  [22] .bss              NOBITS           000000000091ef20  0051ef20
       00000000000310c0  0000000000000000  WA       0     0     32
  [23] .noptrbss         NOBITS           000000000094ffe0  0054ffe0
       0000000000007da0  0000000000000000  WA       0     0     32
  [24] .tbss             NOBITS           0000000000000000  00000000
       0000000000000008  0000000000000000 WAT       0     0     8
  [25] .zdebug_abbrev    PROGBITS         0000000000958000  0051f000
       0000000000000127  0000000000000000           0     0     1
  [26] .zdebug_line      PROGBITS         0000000000958127  0051f127
       000000000003fbd5  0000000000000000           0     0     1
  [27] .zdebug_frame     PROGBITS         0000000000997cfc  0055ecfc
       0000000000017d0a  0000000000000000           0     0     1
  [28] .debug_gdb_s[...] PROGBITS         00000000009afa06  00576a06
       000000000000002a  0000000000000000           0     0     1
  [29] .zdebug_info      PROGBITS         00000000009afa30  00576a30
       0000000000079433  0000000000000000           0     0     1
  [30] .zdebug_loc       PROGBITS         0000000000a28e63  005efe63
       0000000000023dbe  0000000000000000           0     0     1
  [31] .zdebug_ranges    PROGBITS         0000000000a4cc21  00613c21
       0000000000008aad  0000000000000000           0     0     1
  [32] .interp           PROGBITS         0000000000400fe4  00000fe4
       000000000000001c  0000000000000000   A       0     0     1
  [33] .note.go.buildid  NOTE             0000000000400f80  00000f80
       0000000000000064  0000000000000000   A       0     0     4
  [34] .symtab           SYMTAB           0000000000000000  0061c6d0
       000000000002d210  0000000000000018          35   244     8
  [35] .strtab           STRTAB           0000000000000000  006498e0
       00000000000365a1  0000000000000000           0     0     1
Key to Flags:
  W (write), A (alloc), X (execute), M (merge), S (strings), I (info),
  L (link order), O (extra OS processing required), G (group), T (TLS),
  C (compressed), x (unknown), o (OS specific), E (exclude),
  D (mbind), l (large), p (processor specific)

```

As you can see, this binary contains the following sections, among the others:

- .zdebug_abbrev
- .zdebug_line
- .zdebug_frame
- .zdebug_info
- .zdebug_loc
- .zdebug_ranges

These sections are produced by the `gc` compiler with the specific purpose to
improve the debugging process.
The format of these sections is defined by the DWARF Debugging format.
Inspecting the `.debug_info` section will give us the specific version:

```
$ readelf --debug-dump=info server | head
Contents of the .zdebug_info section:

  Compilation Unit @ offset 0x0:
   Length:        0x83 (32-bit)
   Version:       4
   Abbrev Offset: 0x0
   Pointer Size:  8
 <0><b>: Abbrev Number: 1 (DW_TAG_compile_unit)
    <c>   DW_AT_name        : internal/race
    <1a>   DW_AT_language    : 22	(Go)
```

Here we have DWARFv4 debugging info. If you want to know more, the complete
specifications can be found [here](http://dwarfstd.org/Dwarf4Std.php).
Just out of curiosity, the `z` that you find at the start of the section names
means that the section has been compressed.

But if you build your executables passing the options `-s -w` to the linker,
the DWARF information will be stripped away:

```
$ readelf -S server 
There are 27 section headers, starting at offset 0x270:

Section Headers:
  [Nr] Name              Type             Address           Offset
       Size              EntSize          Flags  Link  Info  Align
  [ 0]                   NULL             0000000000000000  00000000
       0000000000000000  0000000000000000           0     0     0
  [ 1] .text             PROGBITS         0000000000401000  00001000
       000000000032c35f  0000000000000000  AX       0     0     32
  [ 2] .plt              PROGBITS         000000000072d360  0032d360
       0000000000000220  0000000000000010  AX       0     0     16
  [ 3] .rodata           PROGBITS         000000000072e000  0032e000
       00000000000a5958  0000000000000000   A       0     0     32
  [ 4] .rela             RELA             00000000007d3958  003d3958
       0000000000000018  0000000000000018   A      11     0     8
  [ 5] .rela.plt         RELA             00000000007d3970  003d3970
       0000000000000318  0000000000000018   A      11     2     8
  [ 6] .gnu.version      VERSYM           00000000007d3ca0  003d3ca0
       000000000000004c  0000000000000002   A      11     0     2
  [ 7] .gnu.version_r    VERNEED          00000000007d3d00  003d3d00
       0000000000000070  0000000000000000   A      10     1     8
  [ 8] .hash             HASH             00000000007d3d80  003d3d80
       00000000000000bc  0000000000000004   A      11     0     8
  [ 9] .shstrtab         STRTAB           0000000000000000  003d3e40
       0000000000000111  0000000000000000           0     0     1
  [10] .dynstr           STRTAB           00000000007d3f60  003d3f60
       0000000000000230  0000000000000000   A       0     0     1
  [11] .dynsym           DYNSYM           00000000007d41a0  003d41a0
       0000000000000390  0000000000000018   A      10     1     8
  [12] .typelink         PROGBITS         00000000007d4540  003d4540
       0000000000001284  0000000000000000   A       0     0     32
  [13] .itablink         PROGBITS         00000000007d57e0  003d57e0
       00000000000007b8  0000000000000000   A       0     0     32
  [14] .gosymtab         PROGBITS         00000000007d5f98  003d5f98
       0000000000000000  0000000000000000   A       0     0     1
  [15] .gopclntab        PROGBITS         00000000007d5fa0  003d5fa0
       000000000010afb8  0000000000000000   A       0     0     32
  [16] .go.buildinfo     PROGBITS         00000000008e1000  004e1000
       0000000000000230  0000000000000000  WA       0     0     16
  [17] .dynamic          DYNAMIC          00000000008e1240  004e1240
       0000000000000120  0000000000000010  WA      10     0     8
  [18] .got.plt          PROGBITS         00000000008e1360  004e1360
       0000000000000120  0000000000000008  WA       0     0     8
  [19] .got              PROGBITS         00000000008e1480  004e1480
       0000000000000008  0000000000000008  WA       0     0     8
  [20] .noptrdata        PROGBITS         00000000008e14a0  004e14a0
       0000000000030e78  0000000000000000  WA       0     0     32
  [21] .data             PROGBITS         0000000000912320  00512320
       000000000000bc10  0000000000000000  WA       0     0     32
  [22] .bss              NOBITS           000000000091df40  0051df40
       00000000000310c0  0000000000000000  WA       0     0     32
  [23] .noptrbss         NOBITS           000000000094f000  0054f000
       0000000000007da0  0000000000000000  WA       0     0     32
  [24] .tbss             NOBITS           0000000000000000  00000000
       0000000000000008  0000000000000000 WAT       0     0     8
  [25] .interp           PROGBITS         0000000000400fe4  00000fe4
       000000000000001c  0000000000000000   A       0     0     1
  [26] .note.go.buildid  NOTE             0000000000400f80  00000f80
       0000000000000064  0000000000000000   A       0     0     4
Key to Flags:
  W (write), A (alloc), X (execute), M (merge), S (strings), I (info),
  L (link order), O (extra OS processing required), G (group), T (TLS),
  C (compressed), x (unknown), o (OS specific), E (exclude),
  D (mbind), l (large), p (processor specific)
```

As you can see, no debugging info sections are there. This is confirmed by the
`file` utility too:

```bash
$ file server
server: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, Go BuildID=gWG4G1efx91jDhIfMBpJ/gwiV9UGqJ7DgiXjLcjWh/slx5duiiso1iTJs2dc4t/ifW5B13fHgj8M06QHszi, stripped
```

It tells us that the binary is **stripped**.
Let's try to run a debug session against a stripped binary:

```
$ dlv exec ./server 
could not launch process: could not open debug info - debuggee must not be built with 'go run' or -ldflags='-s -w', which strip debug info
```

That is not possible at all!
Since stripping an executable is helpful to reduce its size and to reduce
available information for an attacker, what should we do?
Luckily, we can use a core dump, taken from a stripped binary, and run a
`dlv core` session using that core dump and a non-stripped executable.
To do this we must ensure that the non-stripped binary is build exactly from
the same source code.
Recent version of Go may provide reproducible builds under certain conditions.
For debugging purposes, we must ensure that the compiler version is the same
and the source code matches perfectly.
Regarding the source code match prerequisite, this must be true for all the
dependencies of your code, too.
If your using modules, as you should with recent versions of Go, the `go.sum`
will lock your dependencies versions. So, as long as they are still accessible
in the future, you will be able to deterministically build your executable
again. To avoid possible issues, like a taken-down dependency or a transient
failure in the remote repository, you have two options:

- vendoring your dependencies with `go mod vendor`
- use a Go module proxy, like [Athens](https://github.com/gomods/athens)

Finally, let's have a look at the file system path trimming in the executable.

Inspecting the binary with the `strings` utility gives us:

```bash
$ strings server | grep main.go
/home/vagrant/debugging-with-delve/04-core-dumps-and-snapshots/main.go
```

The full path of the local source file ended up in the binary. That's why you
may want to add the `-trimpath` option for your production environment build:

```bash
$ go build -trimpath -gcflags=all='-N -l' -o server
$ strings server | grep main.go
github.com/develersrl/debugging-with-delve/04-core-dumps-and-snapshots/main.go
```

As you can see, the path is now referring to the **remote repository** of the
package.

Now, let's run the executable and try to attach to it with Delve in another
terminal:

```
$ dlv attach $(pidof server)
Type 'help' for list of commands.
(dlv) list main.go:1
Showing github.com/develersrl/debugging-with-delve/04-core-dumps-and-snapshots/main.go:1 (PC: 0x0)
Command failed: open github.com/develersrl/debugging-with-delve/04-core-dumps-and-snapshots/main.go: no such file or directory
```

Sources aren't available. We can tell Delve where to get the sources setting
the configuration option: `substitute-path`:

```
(dlv) config substitute-path github.com/pippolo84/ /home/vagrant/
(dlv) list main.go:1
Showing github.com/develersrl/debugging-with-delve/04-core-dumps-and-snapshots/main.go:1 (PC: 0x0)
     1:	package main
     2:	
     3:	import (
     4:		"fmt"
     5:		"log"
     6:		"net/http"
(dlv) 
```

Cool! Now we know how to unleash the full power of core dump/snapshot debugging
in production environment!