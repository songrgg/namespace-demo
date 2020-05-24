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

## Cgroups
Use cgroups to control the container's CPU and memory usage, because it needs root privilege to update the cgroups, so we need to run with `sudo`.

The uid,gid here is the user ID and group ID mapping to the root user,group in the container, by default they're both 1000.
```shell
$ id
uid=1000(srjiang) gid=1000(srjiang) 
$ sudo NEWROOT=/home/srjiang/Downloads/alpine_root go run exercise05/main.go -uid=1000 -gid=1000
```

In the host, cgroup information can be checked under `/sys/fs/cgroup/cpu/mycontainer/`,
```shell
cat /sys/fs/cgroup/cpu/mycontainer/cgroup.procs
```
The process IDs in the file are the container process and its child process.
