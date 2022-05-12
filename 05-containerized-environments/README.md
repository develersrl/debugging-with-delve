
## Containerized Environments

You might often find yourself needing to debug a process running inside
a container. Perhaps your local development setup requires running multiple
containers with `docker-compose` or you are trying to debug a workload running
in a Kubernetes cluster.
Debugging inside a container introduces a few difficulties that we will go
over in this section.

First, let's build the container we are going to use for this exercise:

```
$ cd debugging-with-delve/05-containerized-environments
$ docker build -t dlv-docker-test .
Sending build context to Docker daemon  17.92kB
Step 1/8 : FROM golang:1.18 AS builder
 ---> 7d1902a99d63
Step 2/8 : WORKDIR /
 ---> Running in 4afa06cf0564
Removing intermediate container 4afa06cf0564
 ---> ec7c5e73abe3
Step 3/8 : COPY main.go /
 ---> 6f33d62be1d8
Step 4/8 : RUN go build -gcflags="all=-N -l" -o server main.go
 ---> Running in 26ddcf6e3792
Removing intermediate container 26ddcf6e3792
 ---> 7941db938320
Step 5/8 : FROM ubuntu:jammy
jammy: Pulling from library/ubuntu
125a6e411906: Pull complete 
Digest: sha256:26c68657ccce2cb0a31b330cb0be2b5e108d467f641c62e13ab40cbec258c68d
Status: Downloaded newer image for ubuntu:jammy
 ---> d2e4e1f51132
Step 6/8 : COPY --from=builder /server /
 ---> 5b55e997ce90
Step 7/8 : EXPOSE 12345/tcp
 ---> Running in 75721cb48554
Removing intermediate container 75721cb48554
 ---> 8e14a1f0599a
Step 8/8 : CMD ["/server"]
 ---> Running in 72749ecc3768
Removing intermediate container 72749ecc3768
 ---> 4417fdd39baf
Successfully built 4417fdd39baf
Successfully tagged dlv-docker-test:latest
```

And then run it:

```
$ docker run -ti --rm -p 12345:12345 --name dlv-docker-test dlv-docker-test
listening on :12345
```

Now in another terminal, we should be able to curl this server:

```
$ curl localhost:12345
Hello, Debugging Workshop!
```

We will need a version of the `dlv` command inside the container to exec.

```
$ go get -d github.com/go-delve/delve/cmd/dlv
$ go build github.com/go-delve/delve/cmd/dlv
$ docker cp ./dlv dlv-docker-test:/dlv
```

Now that we have the `dlv` command inside our container we should be able to
just exec it and attach to the process:

```bash
$ docker exec -ti dlv-docker-test /bin/bash
```

Let's try to attach:

```
root@f7c9ace75770:/# ./dlv attach $(pidof server)
Type 'help' for list of commands.
(dlv)
```

It may happen that you encounter this issue when trying to attach:

```
root@375106c0457d:/# ./dlv attach $(pidof server)
could not attach to pid 1: operation not permitted
```

In that case rerun the container adding the `CAP_SYS_PTRACE` capability:

```bash
docker run --cap-add=SYS_PTRACE -ti --rm -p 12345:12345 --name dlv-docker-test dlv-docker-test
```

After that you should be finally able to attach.

Cool, we successfully attached to the process. Let's set a breakpoint on the
handler:

```
(dlv) b main.handler
Breakpoint 1 set at 0x72cdca for main.handler() .main.go:14
(dlv) c
```

Now, on another terminal, let's curl the server again to trigger the
breakpoint:

```
$ curl localhost:12345
```

The breakpoint fired but we are unable to see the source code:

```
> main.handler() .main.go:14 (hits goroutine(5):1 total:1) (PC: 0x72cdca)
(dlv) list
> main.handler() .main.go:14 (hits goroutine(5):1 total:1) (PC: 0x72cdca)
Command failed: open /main.go: no such file or directory
(dlv)  
```

We can see the function names in the stack trace as well as file and line
numbers for where those functions are defined:

```
(dlv) bt
0  0x000000000072cdca in main.handler
   at .main.go:14
1  0x0000000000715903 in net/http.HandlerFunc.ServeHTTP
   at .usr/local/go/src/net/http/server.go:2084
2  0x0000000000719074 in net/http.serverHandler.ServeHTTP
   at .usr/local/go/src/net/http/server.go:2916
3  0x0000000000714b5c in net/http.(*conn).serve
   at .usr/local/go/src/net/http/server.go:1966
4  0x0000000000719ee5 in net/http.(*Server).Serve.func3
   at .usr/local/go/src/net/http/server.go:3071
5  0x000000000046a421 in runtime.goexit
   at .usr/local/go/src/runtime/asm_amd64.s:1571
(dlv) 
```

This information is contained in the binary. It is usually enough to
manually cross-reference with the source code.
But if we want to improve the debug experience, we just need to make
the sources available to Delve.

Let's copy them into the container:

```bash
$ docker cp main.go dlv-docker-test:/
```

Now query again the server to fire the breakpoint:

```
(dlv) goroutines -with user
  Goroutine 1 - User: .usr/local/go/src/net/fd_unix.go:172 net.(*netFD).accept (0x61b832) [IO wait]
[1 goroutines]

(dlv) goroutine 1 bt
 0  0x000000000043d152 in runtime.gopark
    at .usr/local/go/src/runtime/proc.go:362
 1  0x0000000000435a2a in runtime.netpollblock
    at .usr/local/go/src/runtime/netpoll.go:522
 2  0x0000000000465505 in internal/poll.runtime_pollWait
    at .usr/local/go/src/runtime/netpoll.go:302
 3  0x00000000004d4be8 in internal/poll.(*pollDesc).wait
    at .usr/local/go/src/internal/poll/fd_poll_runtime.go:83
 4  0x00000000004d4c97 in internal/poll.(*pollDesc).waitRead
    at .usr/local/go/src/internal/poll/fd_poll_runtime.go:88
 5  0x00000000004d6c29 in internal/poll.(*FD).Accept
    at .usr/local/go/src/internal/poll/fd_unix.go:614
 6  0x000000000061b832 in net.(*netFD).accept
    at .usr/local/go/src/net/fd_unix.go:172
 7  0x000000000063b3d5 in net.(*TCPListener).accept
    at .usr/local/go/src/net/tcpsock_posix.go:139
 8  0x0000000000639947 in net.(*TCPListener).Accept
    at .usr/local/go/src/net/tcpsock.go:288
 9  0x000000000072c3fc in net/http.(*onceCloseListener).Accept
    at <autogenerated>:1
10  0x0000000000719928 in net/http.(*Server).Serve
    at .usr/local/go/src/net/http/server.go:3039
11  0x0000000000719345 in net/http.(*Server).ListenAndServe
    at .usr/local/go/src/net/http/server.go:2968
12  0x000000000071a856 in net/http.ListenAndServe
    at .usr/local/go/src/net/http/server.go:3222
13  0x000000000072cce9 in main.main
    at .main.go:11
14  0x000000000043cd38 in runtime.main
    at .usr/local/go/src/runtime/proc.go:250
15  0x000000000046a421 in runtime.goexit
    at .usr/local/go/src/runtime/asm_amd64.s:1571

(dlv) goroutine 1 frame 13 list
Goroutine 1 frame 13 at /main.go:11 (PC: 0x72cce9)
     6:		"net/http"
     7:	)
     8:	
     9:	func main() {
    10:		fmt.Println("listening on :12345")
=>  11:		log.Fatal(http.ListenAndServe(":12345", http.HandlerFunc(handler)))
    12:	}
    13:	
    14:	func handler(rw http.ResponseWriter, req *http.Request) {
    15:		fmt.Println("handler ran")
    16:		rw.Write([]byte("Hello, Debugging Workshop!\n"))
(dlv)
```

If you copy the source code to a different path than where it was originally
built you will need to tell Delve where to find the code. You can do that
with the `substitute-path` setting:

```
(dlv) config substitute-path /original/build/path /new/path
```

## Scratch Images

Some folks like to build their docker images with no files other than the
fully statically linked Go binary. Let's build a docker image like that:

```
$ docker build -t dlv-docker-scratch -f Dockerfile.scratch .
Sending build context to Docker daemon  18.94kB
Step 1/8 : FROM golang:1.18 AS builder
 ---> 7d1902a99d63
Step 2/8 : WORKDIR /
 ---> Using cache
 ---> ec7c5e73abe3
Step 3/8 : COPY main.go /
 ---> Using cache
 ---> 6f33d62be1d8
Step 4/8 : RUN CGO_ENABLED=0 go build -gcflags="all=-N -l" -o server main.go
 ---> Running in 0328ec1a9c33
Removing intermediate container 0328ec1a9c33
 ---> 1e38f8871b12
Step 5/8 : FROM scratch
 ---> 
Step 6/8 : COPY --from=builder /server /
 ---> 32b174c9b845
Step 7/8 : EXPOSE 12345/tcp
 ---> Running in dad0ad8a9506
Removing intermediate container dad0ad8a9506
 ---> 333cebdd7434
Step 8/8 : CMD ["/server"]
 ---> Running in a89b81ea3640
Removing intermediate container a89b81ea3640
 ---> d7d5641b7317
Successfully built d7d5641b7317
Successfully tagged dlv-docker-scratch:latest
```

And run it as we did before:

```
$ docker run -ti --rm -p 12345:12345 --name dlv-docker-scratch dlv-docker-scratch
listening on :12345
```

The first problem we will run into is that Delve, by default, dynamically links
against the libc:

```
$ ldd dlv
	linux-vdso.so.1 (0x00007ffd318a3000)
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f3f64ed6000)
	/lib64/ld-linux-x86-64.so.2 (0x00007f3f65106000)
```

So we need to build Delve as a static binary with no dynamic linking:

```
$ CGO_ENABLED=0 go build github.com/go-delve/delve/cmd/dlv
$ file dlv
dlv: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, Go BuildID=INTvTsLdLLD_3fvFmwRN/aHrlFbpkkqG06JPQDTwk/cOLCsCBtH1FHNZUKOSkg/dw3SY6vkTMD4spH3-7dc, not stripped
```

`file` confirms that the binary is statically linked.
Now we can copy it into the container, together with the source code, just like
we did before:

```
$ docker cp dlv dlv-docker-scratch:/
$ docker cp main.go dlv-docker-scratch:/
```

However, we can't exec a shell because there is no shell inside the
container: 

```
$ docker exec -ti dlv-docker-scratch sh
OCI runtime exec failed: exec failed: unable to start container process: exec: "sh": executable file not found in $PATH: unknown
```

To overcome this, we have to exec Delve directly.
Our container is using, among the others, a PID namespace:

```
$ sudo lsns | grep server
4026532206 mnt         1 26610 root             /server
4026532207 uts         1 26610 root             /server
4026532208 ipc         1 26610 root             /server
4026532209 pid         1 26610 root             /server
4026532211 net         1 26610 root             /server
```

so the process we want to attach to will have PID equal to 1. Therefore:

```
$ docker exec -ti dlv-docker-scratch /dlv attach 1
Type 'help' for list of commands.
(dlv) list main.handler
Showing /main.go:14 (PC: 0x7274aa)
     9:	func main() {
    10:		fmt.Println("listening on :12345")
    11:		log.Fatal(http.ListenAndServe(":12345", http.HandlerFunc(handler)))
    12:	}
    13:	
    14:	func handler(rw http.ResponseWriter, req *http.Request) {
    15:		fmt.Println("handler ran")
    16:		rw.Write([]byte("Hello, Debugging Workshop!\n"))
    17:	}
(dlv) 
```

And here we are!

## Sidecar Container

Installing extra tools to container images is rarely a good idea. Usually, you
want slim production images: they are faster and (hopefully) safer.
That's why, in the previous scenarios, we copied the `dlv` executable inside
the *target* container, instead of installing it in the image by default.

But with a bit of knowledge of containerization theory, we can avoid the
executable copy too.

A containerized process is just like any other usual process on the system,
except for the fact that it is executing in a sandboxed environment, managed by
the kernel through namespaces.

Let's start our containerized server:

```
$ docker run -ti --rm -p 12345:12345 --name dlv-docker-scratch dlv-docker-scratch
listening on :12345
```

On another terminal we can see the namespaces related to the newly created
container:

```
$ sudo lsns | grep server
4026532209 mnt         1  8156 root             /server
4026532210 uts         1  8156 root             /server
4026532211 ipc         1  8156 root             /server
4026532212 pid         1  8156 root             /server
4026532214 net         1  8156 root             /server
4026532270 cgroup      1  8156 root             /server
```

We will leverage a **sidecar container** that will share the PID namespace to
debug the *target* container.

An example of such an image can be found in [Dockerfile.sidecar].

Let's build that image:

```
$ docker build . -t dlv-debug-sidecar -f Dockerfile.sidecar
Sending build context to Docker daemon  15.82MB
Step 1/5 : FROM golang:1.18 AS builder
 ---> 7d1902a99d63
Step 2/5 : RUN go install github.com/go-delve/delve/cmd/dlv@v1.8.3
 ---> Using cache
 ---> ee90b8d47087
Step 3/5 : FROM ubuntu:jammy
 ---> d2e4e1f51132
Step 4/5 : COPY --from=builder /go/bin/dlv /
 ---> a0ca91754df0
Step 5/5 : CMD ["bash"]
 ---> Running in fac9639e0123
Removing intermediate container fac9639e0123
 ---> 0081d2c6440e
Successfully built 0081d2c6440e
Successfully tagged dlv-debug-sidecar:latest
```

Now, we can start a new container from that image, sharing the same PID
namespace of the server container and bind mounting the current working
directory where the `dlv` executable and the sources are located.

```
$ docker run -ti --rm --name sidecar_debug --cap-add=SYS_PTRACE --pid container:dlv-docker-scratch -v $(pwd):/srcs dlv-debug-sidecar

root@4d1b7746cb13:/# ps aux | grep server       
root           1  0.0  0.2 706836  5072 pts/0    Ssl+ 08:48   0:00 /server
root          19  0.0  0.0   3468  1440 pts/0    S+   08:58   0:00 grep --color=auto server
```

As expected, the `server` process is visible in the sidecar container too.

Now, let's start Delve:

```
root@3f3e0c63984f:/# ./dlv attach $(pidof server)
Type 'help' for list of commands.
(dlv) 
```

The only thing left to do is to set the correct path for the source code to be
used by Delve:

```
(dlv) config substitute-path "/" /srcs

(dlv) list main.go:1
Showing /main.go:1 (PC: 0x0)
     1:	package main
     2:	
     3:	import (
     4:		"fmt"
     5:		"log"
     6:		"net/http"
```

Excellent! Now we can debug the server as usual!

## Remote Debugging

Sometimes, it is useful to build an image just to start a debug session in a
specific environment.

To do that, you can conveniently use the remote debugging feature, starting
Delve in server mode and attaching a frontend from outside.

An example of such an image can be found in [Dockerfile.remote]

The `CMD` directive in the image start `dlv` in headless mode. Besides, the
image exposes two ports: `8000` to reach the Delve server and `12345` to reach
the application.

Let's build and run it:

```bash
$ docker build -t dlv-docker-remote . -f Dockerfile.remote
Sending build context to Docker daemon  15.82MB
Step 1/10 : FROM golang:1.18 AS builder
 ---> 7d1902a99d63
Step 2/10 : RUN go install github.com/go-delve/delve/cmd/dlv@v1.8.3
 ---> Using cache
 ---> ee90b8d47087
Step 3/10 : WORKDIR /
 ---> Running in 00fb4827ae9f
Removing intermediate container 00fb4827ae9f
 ---> 76fe613d1b7a
Step 4/10 : COPY main.go /
 ---> bcfa4b8f4f37
Step 5/10 : RUN go build -gcflags="all=-N -l" -o server main.go
 ---> Running in 0f2e624789e8
Removing intermediate container 0f2e624789e8
 ---> 219fee8343c6
Step 6/10 : FROM ubuntu:jammy
 ---> d2e4e1f51132
Step 7/10 : COPY --from=builder /go/bin/dlv /
 ---> Using cache
 ---> a0ca91754df0
Step 8/10 : COPY --from=builder /server /
 ---> c85c91f88f01
Step 9/10 : EXPOSE 8000/tcp 12345/tcp
 ---> Running in 9e8094901de6
Removing intermediate container 9e8094901de6
 ---> 067ffff11b37
Step 10/10 : CMD ["/dlv", "--listen=:8000", "--headless=true", "--api-version=2", "exec", "/server"]
 ---> Running in 7faa5629ac91
Removing intermediate container 7faa5629ac91
 ---> 21f0d01c7d06
Successfully built 21f0d01c7d06
Successfully tagged dlv-docker-remote:latest

$ docker run -ti --rm -p 8000:8000 -p 12345:12345 --name dlv-docker-remote dlv-docker-remote
API server listening at: [::]:8000
2022-05-25T09:26:24Z warning layer=rpc Listening for remote connections (connections are not authenticated nor encrypted)
```

On another terminal, we can connect a Delve client to the server listening
inside the container:

```
$ dlv connect localhost:8000 --api-version 2
Type 'help' for list of commands.
(dlv) 
```

Before proceeding, we have to fix the source path.
Note that now we can avoid copying the source code in the container, since the
Delve client will read it from outside.

```
(dlv) config substitute-path / /home/vagrant/debugging-with-delve/05-containerized-environments/
```

As usual, let's set a breakpoint on the HTTP handler and continue execution:

```
(dlv) b main.handler
Breakpoint 1 set at 0x6e050a for main.handler() ./main.go:14
(dlv) c
```

On a third terminal, we can `curl` the server:

```
$ curl localhost:12345
```

and the breakpoint will be triggered:

```
> main.handler() ./main.go:14 (hits goroutine(19):1 total:1) (PC: 0x72cdca)
     9:	func main() {
    10:		fmt.Println("listening on :12345")
    11:		log.Fatal(http.ListenAndServe(":12345", http.HandlerFunc(handler)))
    12:	}
    13:	
=>  14:	func handler(rw http.ResponseWriter, req *http.Request) {
    15:		fmt.Println("handler ran")
    16:		rw.Write([]byte("Hello, Debugging Workshop!\n"))
    17:	}
(dlv) 
```

That's all for debugging in containerized environments!

[Dockerfile.remote]: Dockerfile.remote
[Dockerfile.sidecar]: Dockerfile.sidecar