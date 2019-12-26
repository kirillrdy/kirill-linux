package main

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
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
	expetedPath := path.Join(DistfilesPath, path.Base(url))
	if _, err := os.Stat(expetedPath); os.IsNotExist(err) {
		Cd(DistfilesPath)
		Curl("-O", url)
		Cd(Cwd)
	}
}

func packageVersion(url string) string {
	fileName := path.Base(url)
	//TODO something better in the future
	endings := []string{
		".tar.gz",
		".source.tar.xz", //Hack for firefox
		".tar.xz",
		".zip",
		".tgz",
		".tar.bz2"}

	var result = fileName
	for _, ending := range endings {
		result = strings.TrimSuffix(result, ending)
	}
	return result
}

func extract(url string) {
	fileName := path.Base(url)
	tarballPath := path.Join(DistfilesPath, fileName)
	Tar("xf", tarballPath, "-C", BuildPath)

	extractedPath := path.Join(BuildPath, packageVersion(url))
	Cd(extractedPath)
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
	err = os.MkdirAll(BuildPath, os.ModePerm)
	crash(err)
}

func install(url string, build func(string)) {
	fetch(url)
	extract(url)
	sourceDir := path.Join(BuildPath, packageVersion(url))
	destDir := path.Join(BuildPath, packageVersion(url)+"-package")
	build(destDir)

	//clean up and restore cwd
	Rm("-rf", sourceDir)
	//Rm("-rf", destDir)
	Cd(Cwd)
}

func main() {
	setUpGlobals()

	install("http://ftp.gnu.org/gnu/glibc/glibc-2.30.tar.xz", func(destDir string) {
		Mkdir("build")
		Cd("build")
		DotDotConfigure("--prefix=/usr",
			"--disable-werror",
			"--enable-kernel=3.2",
			"--enable-stack-protector=strong",
			"--with-headers=/usr/include",
			"libc_cv_slibdir=/lib")

		Make("-j10")

		Make("install", "DESTDIR="+destDir)
		//Make("install_root="+destDir, "install")
	})

	install("http://ftp.gnu.org/gnu/binutils/binutils-2.32.tar.xz", func(destDir string) {
		Mkdir("build")
		Cd("build")
		DotDotConfigure("--prefix=/usr",
			"--enable-gold",
			"--enable-ld=default",
			"--enable-plugins",
			"--enable-shared",
			"--disable-werror",
			"--enable-64-bit-bfd",
			"--with-system-zlib")
		Make("-j10")
		Make("install", "DESTDIR="+destDir)
	})

	install("http://ftp.gnu.org/gnu/ncurses/ncurses-6.1.tar.gz", func(destDir string) {
		DotConfigure("--prefix=/usr",
			"--mandir=/usr/share/man",
			"--with-shared",
			"--without-debug",
			"--without-normal",
			"--enable-pc-files",
			"--enable-widec")
		Make("-j10")
		Make("install", "DESTDIR="+destDir)
	})

	install("http://ftp.gnu.org/gnu/bash/bash-5.0.tar.gz", func(destDir string) {
		DotConfigure("--prefix=/usr",
			"--docdir=/usr/share/doc/bash-5.0",
			"--without-bash-malloc",
			"--with-installed-readline")

		Make("-j10")
		Make("install", "DESTDIR="+destDir)
	})
}
