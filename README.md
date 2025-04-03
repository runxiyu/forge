# Lindenii Forge

**Work in progress.**

I tried to reimplement in the [Hare](https://harelang.org) programming
language. It's one of my favorite programming languages, but sadly due to the
lack of multithreading and since it's still a very unstable moving target,
I'm still mainly working on the Go branch.

## Architecture

* Most components are one single daemon written in Hare.
* Because libssh is difficult to use and there aren't many other SSH server
  libraries for C or Hare, we will temporarily use
  [the gliberlabs SSH library for Go](https://github.com/gliderlabs/ssh)
  in a separate process, and communicate via UNIX domain sockets.
