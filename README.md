# Brutescan

Very fast and noisy TCP port scanner. It can scan the whole port range of a device in a second.


## Install

Brutescan is written in go without external dependencies.

Build with

```
go get github.com/asciimoo/brutescan
```


## Usage

```
$ brutescan -h
Usage: ./brutescan [OPTIONS] host/ip
  -pmax uint
        highest target port (default 65535)
  -pmin uint
        lowest target port (default 1)
  -pool uint
        concurrent pool size (default 1000)
  -timeout uint
        connection timeout in milliseconds (default 3000)
  -verbose
        set verbose mode on
```

## Examples

scan localhost:

```
$ brutescan 127.0.0.1
```

scan only high ports with 100 concurrent threads and 10s timeout:

```
$ brutescan -pmin 1024 -pool 100 -timeout 10000 127.0.0.1
```

## Bugs

Bugs or suggestions? Visit the [issue tracker](https://github.com/asciimoo/brutescan/issues).
