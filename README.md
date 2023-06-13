# [Todo] Yet another todo manager.

[![project status][project-status]][project-status]

Todo is an extremely simple toy project, mostly for playing around with new concepts in an app that I actually use.

The biggest reason I use it is because it's purely command line based, only has the features I need, and I can schedule
recurring tasks.

<img src="./demo.gif" />

## Install

### Download a specific release:

You can [view and download releases by version here][releases-url].

### Download the latest release:

- **Linux:** `wget https://github.com/clintjedwards/todo/releases/latest/download/todo`

### Build from source:

You'll need to install [protoc and its associated golang/grpc modules first](https://grpc.io/docs/languages/go/quickstart/)

1. `git clone https://github.com/clintjedwards/todo && cd todo`
2. `make build OUTPUT=/tmp/todo`

The Todo binary comes with a CLI to manage the server as well as act as a client.

## Dev Setup

Todo is setup for easy development. The flag `--dev-mode` flips feature flags such that they enable easy development
features like localhost TLS and easy auth.

### You'll need to install the following first:

To build protocol buffers:

- [protoc](https://grpc.io/docs/protoc-installation/)
- [protoc gen plugins go/grpc](https://grpc.io/docs/languages/go/quickstart/)

### Run from the Makefile

Todo uses flags, env vars, and files to manage configuration (in order of most important). The Makefile already includes all the commands and flags you need to run in dev mode by simply running `make run`.

In case you want to run without the make file simply run:

```bash
export TODO_LOG_LEVEL=debug
go build -o /tmp/$todo
/tmp/todo service start --dev-mode
```

### Editing Protobufs

Todo uses grpc and protobufs to communicate with both plugins and provide an external API. These protobuf
files are located in `/proto`. To compile new protobufs once the original `.proto` files have changed you can use the `make build-protos` command.

### Regenerating Demo Gif

The Gif on the README page uses [vhs](https://github.com/charmbracelet/vhs); a very handy tool that allows you to write a configuration file which will pop out
a gif on the other side.

In order to do this VHS has to run the commands so we must start the server first before we regenerate the gif.

```bash
rm -rf /tmp/todo* # Start with a fresh database
make run # Start the server in dev mode
cd documentation/src/assets
vhs < demo.tape # this will start running commands against the server and output the gif as demo.gif.
```

## Authors

- **Clint Edwards** - [Github](https://github.com/clintjedwards)

This software is provided as-is. It's a hobby project, done in my free time, and I don't get paid for doing it.

[godoc-badge]: https://pkg.go.dev/badge/github.com/clintjedwards/todo
[godoc-url]: https://pkg.go.dev/github.com/clintjedwards/todo
[goreport-badge]: https://goreportcard.com/badge/github.com/clintjedwards/todo
[releases-url]: https://github.com/clintjedwards/todo/releases
[project-status]: https://img.shields.io/badge/Project%20Status-Alpha-orange?style=flat-square
