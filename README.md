[![Go-Curses](https://go-curses.github.io/curses-logo-banner.png)](https://go-curses.github.io)

[![Made with Go](https://img.shields.io/badge/go-v1.16+-blue.svg)](https://golang.org)
[![Go documentation](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/go-curses/cdk)

# CDK - Curses Development Kit

This package provides a very rough [GDK] equivalent for [CTK]. This is not
intended to be a parity of GDK in any way, rather this package simply fulfills
the terminal drawing and basic event systems required by CTK.

Unless you're using CTK, you should really be using [TCell] instead of CDK
directly.

## Notice

This project should not be used for any purpose other than intellectual
curiosity. This status is reflected in the tagged versioning of this `trunk`
branch, v0.1.x, ie: entirely experimental and unfinished in any sense of the
word "done".

## Installing

```
$ go get -u github.com/go-curses/cdk/...
```

## Building

A makefile has been included to assist in the development workflow.

```
$ make help
usage: make [target]

qa targets:
  vet         - run go vet command
  test        - perform all available tests
  cover       - perform all available tests with coverage report

cleanup targets:
  clean       - cleans package and built files
  clean-logs  - cleans *.log from the project

go.mod helpers:
  local       - add go.mod local package replacements
  unlocal     - remove go.mod local package replacements

build targets:
  examples    - builds all examples
  build       - build test for main cdk package
  dev         - build helloworld with profiling

run targets:
  run         - run the dev build (sanely handle crashes)
  profile.cpu - run the dev build and profile CPU
  profile.mem - run the dev build and profile memory
```

## Hello World

An example CDK application demonstrating basic usage and a cdk.Window with customized draw handler.

* see the following source files: [helloworld.go] and [hellowindow.go]

Use the makefile to build the examples.

```
$ make examples 
# cleaning *.log files
# cleaning *.out files
# cleaning pprof files
# cleaning go caches
# cleaning binaries
# building all examples...
#	building helloworld... done.
$ ./helloworld help
NAME:
   helloworld - An example CDK application

USAGE:
   helloworld [global options] command [command options] [arguments...]

VERSION:
   0.0.1

DESCRIPTION:
   Hello World is an example CDK application

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h, --usage  display command-line usage information (default: false)
   --version            display the version (default: false)
```

![helloworld screenshot]

## Application Server

This is something not found in GTK at all and is entirely exclusive to terminal
environments. The idea is simple. Be able to write a terminal interface for
local (/dev/tty) and/or remote connections (ssh).

* see the `appserver` example: [appserver]

### Running the example

To run the example, everything necessary is already included. Start with running
a `make examples`...

```
$ make examples
# cleaning *.log files
# cleaning *.out files
# cleaning pprof files
# cleaning go caches
# cleaning binaries
# building all examples...
#	building appserver... done.
#	building helloworld... done.
```

There should now be an `appserver` binary in the current directory.

```
$ ./appserver help
NAME:
   appserver - run an application server

USAGE:
   appserver [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --listen-port value     sets the port for the server to listen on (default: 2200)
   --listen-address value  sets the address for the server to listen on (default: 0.0.0.0)
   --id-rsa value          sets the path to the server id_rsa file (default: ./examples/appserver/id_rsa)
   --htpasswd value        sets the path to the htpasswd file (default: ./examples/appserver/htpasswd)
   --daemon                start a server daemon instead of a server terminal (default: false)
   --help, -h, --usage     display command-line usage information (default: false)
```

We can see the default settings for the global options to the command. All of
these files are included in the repository and so the defaults should "just
work".

```
$ ./appserver
```

![appserver server-side screenshot]

A version of the [helloworld.go] screen should now appear displaying the time,
roughly to the current second. However, in the background the default port of
`2200` has been opened and is listening for `ssh` connections. Note that it also
has a title at the top: "Server Side".

From a new terminal session, leaving the previous `appserver` running, login
with `ssh` with the username `foo` and the password `bar`, using localhost and
on port `2200`. Note that a second user `bar` exists as well with the password
`foo`.

```
$ ssh -p 2200 foo@localhost
```

![appserver client-side screenshot]

This new session should now be presenting a similar screen as the terminal
server one, with one main difference, the title is "Client Side". This is not to
say that any code is running on the "Client Side"'s shell session, just to say
that this is the "connected via ssh" user interface whereas the "Server Side"
one is the server user interface.

Looking back at the "Server Side" session, it should now report the new client
connection.

![appserver server-side-with-client screenshot]

## Running the tests

There are a handful of sub-packages within the CDK package. The makefile includes
a `make test` option which covers all of these.

```
$ make test
# vetting cdk ... done
# testing cdk ...
=== RUN   FirstTests

  ... (per-test output, trimmed for brevity) ...

--- PASS: OtherTests (0.01s)
PASS
ok      github.com/go-curses/cdk     0.037s
```

Alternatively, [GoConvey] can be used for that delux developer experience.

```
# Install GoConvey first
$ go get github.com/smartystreets/goconvey
 ...
# startup the service, this will open the default browser
$ goconvey
```

## Versioning

The current API is unstable and subject to change dramatically.

## License

This project is licensed under the Apache 2.0 license - see the [LICENSE.md]
file for details.

## Acknowledgments

* Thanks to [TCell] for providing a great starting point for [CDK] and thus
  making [CTK] a plausible reality.

[CTK]: https://github.com/go-curses/ctk
[CDK]: https://github.com/go-curses/cdk
[TCell]: https://github.com/gdamore/tcell
[helloworld.go]: https://github.com/go-curses/cdk/blob/trunk/examples/helloworld/helloworld.go
[hellowindow.go]: https://github.com/go-curses/cdk/blob/trunk/examples/helloworld/hellowindow.go
[appserver]: https://github.com/go-curses/cdk/blob/trunk/examples/appserver
[GoConvey]: https://github.com/smartystreets/goconvey
[LICENSE.md]: https://github.com/go-curses/cdk/blob/trunk/LICENSE.md
[helloworld screenshot]: https://go-curses.github.io/screenshots/cdk-helloworld.png
[appserver server-side screenshot]: https://go-curses.github.io/screenshots/cdk-appserver--server.png
[appserver client-side screenshot]: https://go-curses.github.io/screenshots/cdk-appserver--client.png
[appserver server-side-with-client screenshot]: https://go-curses.github.io/screenshots/cdk-appserver--server-with-client.png
