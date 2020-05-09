# Linux namespace Golang experiments
This project is to understand the Linux namespace: UID, PID, UTS, Mount, Networking namespaces by Golang experiments,
they're separated by several exercises.
For the detailed explanation, please read [Linux namespace in Go - Part 1, UTS and PID](https://songrgg.github.io/programming/linux-namespace-part01-uts-pid/)

## UTS
```shell
sudo go run exercise01/main.go
```

## PID
```shell
sudo go run exercise02/main.go
```

## UID & Mount
```shell
go run exercise03/main.go
```

## Mount a new root filesystem
First download alpine root filesystem from https://alpinelinux.org/downloads/ and get the path,
```shell
NEWROOT=~/Downloads/alpine_root go run exercise04/main.go
```
