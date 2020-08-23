package shell

import (
	"bytes"
	"github.com/kirillrdy/kirill-linux/config"
	"log"
	"os"
	"os/exec"
	"runtime"
)

//TODO have some sort of same thing
func crash(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func ExecInteractive(cmd string, args ...string) {
	if config.Verbose {
		log.Printf("%s %s", cmd, args)
	}
	command := exec.Command(cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	err := command.Run()
	crash(err)
}

func Exec(cmd string, args ...string) {
	if config.Verbose {
		log.Printf("%s %s", cmd, args)
	}
	command := exec.Command(cmd, args...)
	var buffer bytes.Buffer
	command.Stdout = &buffer
	command.Stderr = &buffer
	err := command.Run()
	if err != nil {
		log.Print(buffer.String())
	}
	crash(err)
}

func Curl(args ...string) {
	Exec("curl", args...)
}

func Mkdir(args ...string) {
	Exec("mkdir", args...)
}

func Tar(args ...string) {
	Exec("tar", args...)
}

// This is a hard choice to make, but
// for now on FreeBSD we have to call gmake
func Make(args ...string) {
	if runtime.GOOS == "linux" {
		Exec("make", args...)
	} else {
		Exec("gmake", args...)
	}
}

func Rm(args ...string) {
	Exec("rm", args...)
}

func Mv(args ...string) {
	Exec("mv", args...)
}

func Ln(args ...string) {
	Exec("ln", args...)
}

func DotConfigure(args ...string) {
	Exec("./configure", args...)
}

func DotDotConfigure(args ...string) {
	Exec("../configure", args...)
}

func Cd(dir string) {
	if config.Verbose {
		log.Printf("cd %s", dir)
	}
	os.Chdir(dir)
}
