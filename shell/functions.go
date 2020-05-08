package shell

import (
	"bytes"
	"log"
	"os"
	"os/exec"
)

//TODO have some sort of same thing
func crash(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func ExecInteractive(cmd string, args ...string) {
	log.Printf("%s %s", cmd, args)
	command := exec.Command(cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	err := command.Run()
	crash(err)
}

func Exec(cmd string, args ...string) {
	log.Printf("%s %s", cmd, args)
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

func Make(args ...string) {
	Exec("make", args...)
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
	log.Printf("cd %s", dir)
	os.Chdir(dir)
}
