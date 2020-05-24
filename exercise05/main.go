package main

import (
	"flag"
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

func addProcessToCgroup(filepath string, pid int) {
	file, err := os.OpenFile(filepath, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	if _, err := file.WriteString(fmt.Sprintf("%d", pid)); err != nil {
		fmt.Println("failed to setup cgroup for the container: ", err)
		os.Exit(1)
	}
}

func cgroupSetup(pid int) {
	for _, c := range []string{"cpu", "memory"} {
		cpath := fmt.Sprintf("/sys/fs/cgroup/%s/mycontainer/", c)
		if err := os.MkdirAll(cpath, 0644); err != nil {
			fmt.Println("failed to create cpu cgroup for my container: ", err)
			os.Exit(1)
		}
		addProcessToCgroup(cpath+"cgroup.procs", pid)
	}
}

func main() {
	uidPtr := flag.Int("uid", 1000, "user ID the container will run as")
	gidPtr := flag.Int("gid", 1000, "group ID the container will run as")
	flag.Parse()

	cmd := reexec.Command("alpine_shell")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      *uidPtr, // use non-root user to run as root
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      *gidPtr, // use non-root group to run as root
				Size:        1,
			},
		},
		Credential: &syscall.Credential{
			Uid: 0,
			Gid: 0,
		},
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("failed to start the command: ", err)
		os.Exit(1)
	}

	// cgroup setup
	cgroupSetup(cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error running the exec.Command - %s\n", err)
		os.Exit(1)
	}
}
