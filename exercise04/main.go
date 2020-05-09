package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/docker/docker/pkg/reexec"
)

func init() {
	reexec.Register("alpine_shell", namespaceInit)

	// if it's the child process
	if reexec.Init() {
		os.Exit(0)
	}
}

func namespaceInit() {
	newRoot := os.Getenv("NEWROOT")
	putOld := "/old_root"
	// 1. mount alpine root file system as a mountpoint, then it can be used to pivot_root
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND, ""); err != nil {
		fmt.Println("failed to mount new root filesystem: ", err)
		os.Exit(1)
	}

	if err := syscall.Mkdir(newRoot+putOld, 0700); err != nil {
		fmt.Println("failed to mkdir: ", err)
		os.Exit(1)
	}

	if err := syscall.PivotRoot(newRoot, newRoot+putOld); err != nil {
		fmt.Println("failed to pivot root: ", err)
		os.Exit(1)
	}

	if err := syscall.Chdir("/"); err != nil {
		fmt.Println("failed to chdir to /: ", err)
		os.Exit(1)
	}

	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		fmt.Println("failed to mount /proc: ", err)
		os.Exit(1)
	}

	// unmount the old root filesystem
	if err := syscall.Unmount(putOld, syscall.MNT_DETACH); err != nil {
		fmt.Println("failed to unmount the old root filesystem: ", err)
		os.Exit(1)
	}

	if err := os.RemoveAll(putOld); err != nil {
		fmt.Println("failed to remove old root filesystem: ", err)
		os.Exit(1)
	}

	namespaceRun()
}

func namespaceRun() {
	cmd := exec.Command("/bin/sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("failed to run the command: ", err)
		os.Exit(1)
	}
}

func main() {
	cmd := reexec.Command("alpine_shell")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running the exec.Command - %s\n", err)
		os.Exit(1)
	}
}
