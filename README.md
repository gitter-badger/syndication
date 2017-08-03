# Syndication - An extensible news aggregation server

## Features
* JSON REST API
* Unix socket based Administration API

## Planned Features
* Plugin system

## Building

```
$ mkdir syndication-build && cd syndication-build
$ export GOPATH=$(pwd)
$ mkdir src bin pkg
$ go get github.com/varddum/syndication
$ cd srg/github.com/varddum/syndication
$ go build
```
