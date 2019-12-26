package main

import (
	"log"
	"os"
	"os/exec"
	"path"
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

func fetch(url string) {
	Cd(DistfilesPath)
	Curl("-O", url)
	Cd(Cwd)
}

func extract(url string) {
	fileName := path.Base(url)
	tarballPath := path.Join(DistfilesPath, fileName)
	Tar("xf", tarballPath, "-C", BuildPath)
}

// You can think of this as root directory for everything
var Cwd string
var DistfilesPath string
var BuildPath string

func setUpGlobals() {

	var err error
	Cwd, err = os.Getwd()
	crash(err)

	DistfilesPath = path.Join(Cwd, "distfiles")
	err = os.MkdirAll(DistfilesPath, os.ModePerm)
	crash(err)

	BuildPath = path.Join(Cwd, "build")
	os.RemoveAll(BuildPath)
	err = os.MkdirAll(BuildPath, os.ModePerm)
	crash(err)
}

func installGlibc() {
	url := "http://ftp.gnu.org/gnu/glibc/glibc-2.30.tar.xz"
	fetch(url)
	extract(url)

	extractedDirName := "glibc-2.30"
	workDir := path.Join(BuildPath, extractedDirName)
	destDir := path.Join(BuildPath, extractedDirName+"-package")

	Cd(workDir)
	Mkdir("build")
	Cd("build")
	DotDotConfigure("--prefix=/usr",
		"--disable-werror",
		"--enable-kernel=3.2",
		"--enable-stack-protector=strong",
		"--with-headers=/usr/include",
		"libc_cv_slibdir=/lib")

	Make("-j10")
	Make("install_root="+destDir, "install")

}

func main() {
	setUpGlobals()

	installGlibc()
	Cd(Cwd)

	Curl("-O", "http://ftp.gnu.org/gnu/binutils/binutils-2.32.tar.xz")
	Rm("-rfv", "binutils-2.32")
	Tar("xf", "binutils-2.32.tar.xz")
	Cd("binutils-2.32")
	Mkdir("build")
	Cd("build")

	newRoot := "/home/kirillvr/newroot"

	DotDotConfigure("--prefix=/usr",
		"--enable-gold",
		"--enable-ld=default",
		"--enable-plugins",
		"--enable-shared",
		"--disable-werror",
		"--enable-64-bit-bfd",
		"--with-system-zlib")
	Make("-j10")
	Make("install", "DESTDIR="+newRoot)
	Cd("../..")

	Curl("-O", "http://ftp.gnu.org/gnu/ncurses/ncurses-6.1.tar.gz")
	Rm("-rfv", "ncurses-6.1")
	Tar("xf", "ncurses-6.1.tar.gz")
	Cd("ncurses-6.1")
	DotConfigure("--prefix=/usr",
		"--mandir=/usr/share/man",
		"--with-shared",
		"--without-debug",
		"--without-normal",
		"--enable-pc-files",
		"--enable-widec")
	Make("-j10")
	Make("install", "DESTDIR="+newRoot)
	Cd("..")

	Curl("-O", "http://ftp.gnu.org/gnu/bash/bash-5.0.tar.gz")
	Rm("-rfv", "bash-5.0")
	Tar("xf", "bash-5.0.tar.gz")
	Cd("bash-5.0")
	DotConfigure("--prefix=/usr",
		"--docdir=/usr/share/doc/bash-5.0",
		"--without-bash-malloc",
		"--with-installed-readline")

	Make("-j10")
	Make("install", "DESTDIR="+newRoot)
	Cd("..")
}
