# Lindenii Forge

**Work in progress.**

This is the new implementation in the [Hare](https://harelang.org) programming
language.

## Architecture

* Most components are one single daemon written in Hare.
* Because libssh is difficult to use and there aren't many other SSH server
  libraries for C or Hare, we will temporarily use
  [the gliberlabs SSH library for Go](https://github.com/gliderlabs/ssh)
  in a separate process, and communicate via UNIX domain sockets.
