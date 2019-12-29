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

// Note it also adds \n
func appendToFile(fileName string, items ...string) {
	log.Printf("Appending %v to %s\n", items, fileName)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	crash(err)

	defer file.Close()
	_, err = file.WriteString(strings.Join(items, "\n") + "\n")
	crash(err)
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

func curl(args ...string) {
	execCmd("curl", args...)
}

func mkdir(args ...string) {
	execCmd("mkdir", args...)
}

func tar(args ...string) {
	execCmd("tar", args...)
}

func make(args ...string) {
	execCmd("make", args...)
}

func rm(args ...string) {
	execCmd("rm", args...)
}

func mv(args ...string) {
	execCmd("mv", args...)
}

func ln(args ...string) {
	execCmd("ln", args...)
}

func dotConfigure(args ...string) {
	execCmd("./configure", args...)
}

func dotDotConfigure(args ...string) {
	execCmd("../configure", args...)
}

func cd(dir string) {
	log.Printf("cd %s", dir)
	os.Chdir(dir)
}

func fetch(url string) {
	cd(DistfilesPath)
	curl("-O", "-L", url)
	cd(Cwd)
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

	expetedTarPath := path.Join(DistfilesPath, path.Base(url))
	if _, err := os.Stat(expetedTarPath); os.IsNotExist(err) {
		fetch(url)
	}

	extractedSourcePath := path.Join(BuildPath, packageVersion(url))
	if _, err := os.Stat(extractedSourcePath); os.IsNotExist(err) {
		tarballPath := path.Join(DistfilesPath, path.Base(url))
		tar("xf", tarballPath, "-C", BuildPath)
	}

	cd(extractedSourcePath)
}

// Cwd we remeber cwd in order to get back here
var Cwd string

// DistfilesPath is where all distfiles are stored
var DistfilesPath string

// BuildPath is where all the work is done
var BuildPath string

// PkgPath is where binary packages get stored
var PkgPath string

// InstallPrefix prefix is where we are all going to install things for new system
var InstallPrefix string

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

	InstallPrefix = path.Join(Cwd, "newroot")
	err = os.MkdirAll(InstallPrefix, os.ModePerm)
	crash(err)

}

func installSimple(url string) {
	installConfigure(url, func() {
		dotConfigure("--prefix=/usr")
	})
}

func installConfigure(url string, configure func()) {
	installBuildInstall(url, func() {
		configure()
		//TODO detect 8
		make("-j8")
	}, func(destDir string) {
		make("install", "DESTDIR="+destDir)
	})
}

// stupid name, but what can you do
func installBuildInstall(url string, build func(), install func(string)) {
	tarBall := path.Join(PkgPath, packageVersion(url)+".tar.xz")

	if _, err := os.Stat(tarBall); os.IsNotExist(err) {
		extract(url)
		sourceDir := path.Join(BuildPath, packageVersion(url))

		build()

		// Part of packaging
		destDir := path.Join(BuildPath, packageVersion(url)+"-package")
		install(destDir)
		tar("cf", tarBall, "-C", destDir, ".")
		cd(Cwd)
		rm("-rf", destDir)

		rm("-rf", sourceDir)

	}

	//TODO also dont do this if its already installed eg need some way of tracking those
	//TODO replace with desired prefix
	tar("xf", tarBall, "-C", InstallPrefix)
}

func main() {
	setUpGlobals()

	installBuildInstall("http://ftp.gnu.org/gnu/glibc/glibc-2.30.tar.xz", func() {
		mkdir("build")
		cd("build")
		dotDotConfigure("--prefix=/usr",
			"--disable-werror",
			"--enable-kernel=3.2",
			"--enable-stack-protector=strong",
			"--with-headers=/usr/include",
			"libc_cv_slibdir=/lib")
		make("-j8")
	}, func(destDir string) {
		make("install", "DESTDIR="+destDir)
		cd(destDir)
		//TODO need a better longterm solution
		ln("-s", "lib", "lib64")
	})

	installSimple("https://zlib.net/zlib-1.2.11.tar.xz")
	installSimple("ftp://ftp.astron.com/pub/file/file-5.37.tar.gz")
	installSimple("http://ftp.gnu.org/gnu/readline/readline-8.0.tar.gz")

	installConfigure("http://ftp.gnu.org/gnu/binutils/binutils-2.32.tar.xz", func() {
		mkdir("build")
		cd("build")
		dotDotConfigure("--prefix=/usr",
			"--enable-gold",
			"--enable-ld=default",
			"--enable-plugins",
			"--enable-shared",
			"--disable-werror",
			"--enable-64-bit-bfd",
			"--with-system-zlib")
	})

	installConfigure("https://pkg-config.freedesktop.org/releases/pkg-config-0.29.2.tar.gz", func() {
		dotConfigure("--prefix=/usr",
			"--with-internal-glib",
			"--disable-host-tool")
	})

	installConfigure("http://ftp.gnu.org/gnu/gcc/gcc-9.2.0/gcc-9.2.0.tar.xz", func() {
		//TODO less hardcoded versions
		extract("http://ftp.gnu.org/gnu/gmp/gmp-6.1.2.tar.xz")
		extract("https://www.mpfr.org/mpfr-4.0.2/mpfr-4.0.2.tar.xz")
		extract("https://ftp.gnu.org/gnu/mpc/mpc-1.1.0.tar.gz")

		//TODO less
		cd("../gcc-9.2.0")
		mv("../gmp-6.1.2", "gmp")
		mv("../mpfr-4.0.2", "mpfr")
		mv("../mpc-1.1.0", "mpc")

		mkdir("build")
		cd("build")
		dotDotConfigure("--prefix=/usr",
			"--enable-languages=c,c++",
			"--disable-multilib",
			"--disable-bootstrap",
			"--with-system-zlib")
	})

	installConfigure("http://ftp.gnu.org/gnu/ncurses/ncurses-6.1.tar.gz", func() {
		dotConfigure("--prefix=/usr",
			"--mandir=/usr/share/man",
			"--with-shared",
			"--without-debug",
			"--without-normal",
			"--enable-pc-files",
			"--enable-widec")
	})

	installConfigure("http://ftp.gnu.org/gnu/bash/bash-5.0.tar.gz", func() {
		dotConfigure("--prefix=/usr",
			"--docdir=/usr/share/doc/bash-5.0",
			"--without-bash-malloc",
			"--with-installed-readline")

	})

	installSimple("http://ftp.gnu.org/gnu/coreutils/coreutils-8.31.tar.xz")
	installSimple("https://github.com/vim/vim/archive/v8.1.1846/vim-8.1.1846.tar.gz")

	installBuildInstall("https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-5.4.6.tar.xz", func() {

		// zfs has very slow configure time, so disabling it until i get to zfs on root
		enableZFS := false

		if enableZFS {
			extract("https://github.com/zfsonlinux/zfs/releases/download/zfs-0.8.2/zfs-0.8.2.tar.gz")
			dotConfigure("--enable-linux-builtin")
			make("-j8")
			execCmd("./copy-builtin", "../linux-5.4.6")
			cd("../linux-5.4.6")
			rm("-rfv", "../zfs-0.8.2")
		}
		make("defconfig")
		if enableZFS {
			appendToFile(".config", "CONFIG_ZFS=y")
		}
		appendToFile(".config",
			"CONFIG_CMDLINE_BOOL=y",
			"CONFIG_CMDLINE=\"rootwait root=/dev/sdc2 init=/usr/bin/bash\"",
			"CONFIG_DRM_NOUVEAU=y",
		)

		make("-j8")
	}, func(destDir string) {
		//TODO dont forget modules as well
		mkdir("-p", path.Join(destDir, "/boot/efi/EFI/boot"))
		mv("arch/x86/boot/bzImage", path.Join(destDir, "/boot/efi/EFI/boot/bootx64.efi"))
	})

}

// TODO before extraction check for clashes
