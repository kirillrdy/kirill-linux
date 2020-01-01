package main

import (
	"os"
	"os/exec"
)

func crash(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	//TODO also print stdout etc
	err := exec.Command("/bin/mount", "-o", "remount,rw", "/").Run()
	crash(err)
	err = exec.Command("/bin/mount", "-t", "proc", "/proc").Run()
	crash(err)
	cmd := exec.Command("/sbin/agetty", "tty1", "9600")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	crash(err)
}
