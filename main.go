package main

import (
	"bytes"
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

func Mv(args ...string) {
	execCmd("mv", args...)
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
		Curl("-O", "-L", url)
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
var PkgPath string

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

	PkgPath = path.Join(Cwd, "package")
	err = os.MkdirAll(PkgPath, os.ModePerm)
	crash(err)
}

func installSimple(url string) {
	install(url, func(destDir string) {
		DotConfigure("--prefix=/usr")
		Make("-j8")
		Make("install", "DESTDIR="+destDir)
	})
}

func install(url string, build func(string)) {
	tarBall := path.Join(PkgPath, packageVersion(url)+".tar.xz")

	if _, err := os.Stat(tarBall); os.IsNotExist(err) {
		fetch(url)
		extract(url)
		sourceDir := path.Join(BuildPath, packageVersion(url))
		destDir := path.Join(BuildPath, packageVersion(url)+"-package")
		build(destDir)
		Tar("cf", tarBall, "-C", destDir, ".")
		Rm("-rf", sourceDir)
		Rm("-rf", destDir)
		Cd(Cwd)
	}

	//TODO also dont do this if its already installed eg need some way of tracking those
	//TODO replace with desired prefix
	Tar("xf", tarBall, "-C", "/home/kirillvr/newroot")
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
	})

	installSimple("https://zlib.net/zlib-1.2.11.tar.xz")
	installSimple("ftp://ftp.astron.com/pub/file/file-5.37.tar.gz")
	installSimple("http://ftp.gnu.org/gnu/readline/readline-8.0.tar.gz")

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

	install("https://pkg-config.freedesktop.org/releases/pkg-config-0.29.2.tar.gz", func(destDir string) {
		DotConfigure("--prefix=/usr",
			"--with-internal-glib",
			"--disable-host-tool")
		Make("-j10")
		Make("install", "DESTDIR="+destDir)
	})

	install("http://ftp.gnu.org/gnu/gcc/gcc-9.2.0/gcc-9.2.0.tar.xz", func(destDir string) {
		//TODO less hardcoded versions
		fetch("http://ftp.gnu.org/gnu/gmp/gmp-6.1.2.tar.xz")
		extract("http://ftp.gnu.org/gnu/gmp/gmp-6.1.2.tar.xz")
		fetch("https://www.mpfr.org/mpfr-4.0.2/mpfr-4.0.2.tar.xz")
		extract("https://www.mpfr.org/mpfr-4.0.2/mpfr-4.0.2.tar.xz")
		fetch("https://ftp.gnu.org/gnu/mpc/mpc-1.1.0.tar.gz")
		extract("https://ftp.gnu.org/gnu/mpc/mpc-1.1.0.tar.gz")

		Cd("../gcc-9.2.0")
		Mv("../gmp-6.1.2", "gmp")
		Mv("../mpfr-4.0.2", "mpfr")
		Mv("../mpc-1.1.0", "mpc")

		Mkdir("build")
		Cd("build")
		DotDotConfigure("--prefix=/usr",
			"--enable-languages=c,c++",
			"--disable-multilib",
			"--disable-bootstrap",
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

	installSimple("https://github.com/vim/vim/archive/v8.1.1846/vim-8.1.1846.tar.gz")

}

// package things
// extract make bits
// build gcc
// before extraction check for clashes
