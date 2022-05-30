# Debugging Go applications with Delve

<div align="center">
<img width="640" src="assets/github.jpg">
</div>

<br />

### Workshop Objective

The workshop aims to illustrate the basics of Delve, the Go debugger.

### Who It Is For

Developers with a previous experience with Go that would like to improve their
Go debugging skills.

### Requirements

Basic experience with Go

### Environment Setup

For the majority of the exercises what you need is to:

- [Install Go][install-go]
- [Install Delve][install-delve]

To follow the workshop contents and try all the examples, you can run the
`Vagrantfile` included to provision a Linux VM.
This way, you will get rid of all the possible annoyances that may come up
configuring all the tools.
To do that you will need to:

- [Install VirtualBox][install-virtualbox]
- [Install Vagrant][install-vagrant]
- Run `vagrant up` from the repo root to provision the VM
- Get a shell on the VM by running `vagrant ssh`

### Outline

The workshop is broken down into sections. We will go through these individually
during the class:

- [First Contact With Delve][01]
- [Taming Concurrency][02]
- [Debugging Sessions][03]
- [Core Dumps and Snapshots][04]
- [Containerized Environments][05]
- [Debuggers Under The Hood][06]
- [References and Further Readings][07]

Please ask questions and interact freely with the speaker during the workshop!

### Contacts

email: fabio [at] develer [dot] com

You can find me on [Twitter] too!

[install-go]: http://golang.org/dl
[install-delve]: https://github.com/go-delve/delve/tree/master/Documentation/installation
[install-virtualbox]: https://www.virtualbox.org/wiki/Downloads
[install-vagrant]: https://www.vagrantup.com/downloads

[Twitter]: https://twitter.com/Pippolo84


[01]: https://github.com/develersrl/debugging-with-delve/blob/main/01-first-contact-with-delve
[02]: https://github.com/develersrl/debugging-with-delve/blob/main/02-taming-concurrency
[03]: https://github.com/develersrl/debugging-with-delve/blob/main/03-debugging-sessions
[04]: https://github.com/develersrl/debugging-with-delve/blob/main/04-core-dumps-and-snapshots
[05]: https://github.com/develersrl/debugging-with-delve/blob/main/05-containerized-environments
[06]: https://github.com/develersrl/debugging-with-delve/blob/main/06-debuggers-under-the-hood
[07]: https://github.com/develersrl/debugging-with-delve/blob/main/07-references