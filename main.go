package main

import (
	"log"
	"os"
	"os/exec"
)

func crash(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func execCmd(cmd string, args ...string) {
	log.Printf("%s %s", cmd, args)
	err := exec.Command(cmd, args...).Run()
	crash(err)
}

func Curl(args ...string) {
	execCmd("curl", args...)
}

func Mkdir(args ...string) {
	execCmd("mkdir", args...)
}

func Tar(args ...string) {
	execCmd("tar", args...)
}

func Make(args ...string) {
	execCmd("make", args...)
}

func Rm(args ...string) {
	execCmd("rm", args...)
}

func DotConfigure(args ...string) {
	execCmd("./configure", args...)
}

func DotDotConfigure(args ...string) {
	execCmd("../configure", args...)
}

func Cd(dir string) {
	log.Printf("cd %s", dir)
	os.Chdir(dir)
}

func main() {
	newRoot := "/home/kirillvr/newroot"

	Rm("-rf", "glibc-2.30")
	Curl("-O", "http://ftp.gnu.org/gnu/glibc/glibc-2.30.tar.xz")
	Tar("xf", "glibc-2.30.tar.xz")
	Cd("glibc-2.30")
	Mkdir("build")
	Cd("build")
	DotDotConfigure("--prefix=/usr",
		"--disable-werror",
		"--enable-kernel=3.2",
		"--enable-stack-protector=strong",
		"--with-headers=/usr/include",
		"libc_cv_slibdir=/lib")

	Make("-j10")
	Make("install_root="+newRoot, "install")
	Cd("../..")
}
