package main

import (
	"flag"
	. "github.com/kirillrdy/kirill-linux/shell"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

//TODO have some sort of same thing
func crash(err error) {
	if err != nil {
		log.Panic(err)
	}
}

//Note it also adds a new line
func createFile(fileName string, content string) {
	log.Printf("Creating file %s with content %s\n", fileName, content)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0644)
	crash(err)

	defer file.Close()
	_, err = file.WriteString(content + "\n")
	crash(err)

}

// Note it also adds \n for each item it writes
func appendToFile(fileName string, items ...string) {
	log.Printf("Appending %v to %s\n", items, fileName)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	crash(err)

	defer file.Close()
	_, err = file.WriteString(strings.Join(items, "\n") + "\n")
	crash(err)
}

func execCmdInteractive(cmd string, args ...string) {
	log.Printf("%s %s", cmd, args)
	command := exec.Command(cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	err := command.Run()
	crash(err)
}

func fetch(url string) {
	log.Printf("fetching %s\n", url)
	Cd(DistfilesPath)

	// Get the data
	resp, err := http.Get(url)
	crash(err)
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(path.Base(url))
	crash(err)
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	crash(err)

	Cd(Cwd)
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
		Tar("xf", tarballPath, "-C", BuildPath)
	}

	Cd(extractedSourcePath)
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
		DotConfigure("--prefix=/usr")
	})
}

//TODO detect
const NumberOfMakeJobs = "-j8"

func installConfigure(url string, configure func()) {
	installBuildInstall(url, func() {
		configure()

		Make(NumberOfMakeJobs)
	}, func(destDir string) {
		Make("install", "DESTDIR="+destDir)
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
		Tar("cf", tarBall, "-C", destDir, ".")
		Cd(Cwd)
		Rm("-rf", destDir)
		Rm("-rf", sourceDir)
	}

	//TODO also dont do this if its already installed eg need some way of tracking those
	//TODO replace with desired prefix
	Tar("xf", tarBall, "-C", InstallPrefix)
}

func main() {
	setUpGlobals()

	enterChroot := flag.Bool("c", false, "enter chroot")
	flag.Parse()

	Mkdir("-p", path.Join(InstallPrefix, "etc"))
	Mkdir("-p", path.Join(InstallPrefix, "tmp"))
	Mkdir("-p", path.Join(InstallPrefix, "dev"))
	Mkdir("-p", path.Join(InstallPrefix, "sys"))
	Mkdir("-p", path.Join(InstallPrefix, "run"))
	Mkdir("-p", path.Join(InstallPrefix, "root"))
	Mkdir("-p", path.Join(InstallPrefix, "proc"))

	createFile(path.Join(InstallPrefix, "etc/passwd"), "root::0:0:root:/root:/bin/bash\n")
	createFile(path.Join(InstallPrefix, "etc/group"),
		`
root:x:0:
bin:x:1:
sys:x:2:
kmem:x:3:
tty:x:4:
tape:x:5:
daemon:x:6:
floppy:x:7:
disk:x:8:
lp:x:9:
dialout:x:10:
audio:x:11:
video:x:12:
utmp:x:13:
usb:x:14:
`)

	//TODO need some sort of createFile rather than append, so that doesnt do it if file is already
	// there
	createFile(path.Join(InstallPrefix, "etc/fstab"), `
# Begin /etc/fstab

# file system  mount-point  type     options             dump  fsck
#                                                              order

proc           /proc        proc     nosuid,noexec,nodev 0     0
sysfs          /sys         sysfs    nosuid,noexec,nodev 0     0
devpts         /dev/pts     devpts   gid=5,mode=620      0     0
tmpfs          /run         tmpfs    defaults            0     0
devtmpfs       /dev         devtmpfs mode=0755,nosuid    0     0

# End /etc/fstab
  `)

	linuxKernelSourcesURL := "https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-5.4.6.tar.xz"

	installBuildInstall("http://ftp.gnu.org/gnu/glibc/glibc-2.30.tar.xz", func() {
		Mkdir("build")
		Cd("build")
		DotDotConfigure("--prefix=/usr",
			"--disable-werror",
			"--enable-kernel=3.2",
			"--enable-stack-protector=strong",
			"--with-headers=/usr/include",
			"libc_cv_slibdir=/lib")
		Make(NumberOfMakeJobs)
	}, func(destDir string) {
		Make("install", "DESTDIR="+destDir)
		Cd(destDir)
		//TODO need a better longterm solution
		Ln("-s", "lib", "lib64")
		ioutil.WriteFile("etc/ld.so.conf", []byte("/usr/local/lib\n/opt/lib\n"), os.ModePerm)
		nssContent := `
# Begin /etc/nsswitch.conf

passwd: files
group: files
shadow: files

hosts: files dns
networks: files

protocols: files
services: files
ethers: files
rpc: files

# End /etc/nsswitch.conf
    `
		ioutil.WriteFile("etc/nsswitch.conf", []byte(nssContent), os.ModePerm)
	})

	installSimple("https://zlib.net/zlib-1.2.11.tar.xz")
	installSimple("http://ftp.astron.com/pub/file/file-5.37.tar.gz")
	installSimple("http://ftp.gnu.org/gnu/readline/readline-8.0.tar.gz")
	//m4 skipping for now
	//bc skipping as well

	installConfigure("http://ftp.gnu.org/gnu/binutils/binutils-2.32.tar.xz", func() {
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
	})

	installConfigure("http://ftp.gnu.org/gnu/gcc/gcc-9.2.0/gcc-9.2.0.tar.xz", func() {
		//TODO less hardcoded versions
		extract("http://ftp.gnu.org/gnu/gmp/gmp-6.1.2.tar.xz")
		extract("https://www.mpfr.org/mpfr-4.0.2/mpfr-4.0.2.tar.xz")
		extract("https://ftp.gnu.org/gnu/mpc/mpc-1.1.0.tar.gz")

		//TODO less
		Cd("../gcc-9.2.0")
		Mv("../gmp-6.1.2", "gmp")
		Mv("../mpfr-4.0.2", "mpfr")
		Mv("../mpc-1.1.0", "mpc")

		Mkdir("build")
		Cd("build")
		DotDotConfigure("--prefix=/usr",
			"--enable-languages=c,c++,go",
			"--disable-multilib",
			"--disable-bootstrap",
			"--with-system-zlib")
	})

	installConfigure("https://pkg-config.freedesktop.org/releases/pkg-config-0.29.2.tar.gz", func() {
		DotConfigure("--prefix=/usr",
			"--with-internal-glib",
			"--disable-host-tool")
	})

	installConfigure("http://ftp.gnu.org/gnu/grep/grep-3.3.tar.xz", func() {
		DotConfigure("--prefix=/usr", "--bindir=/bin", "--disable-perl-regexp")
	})

	installConfigure("http://ftp.gnu.org/gnu/ncurses/ncurses-6.1.tar.gz", func() {
		DotConfigure("--prefix=/usr",
			"--mandir=/usr/share/man",
			"--with-shared",
			"--without-debug",
			"--without-normal",
			"--enable-pc-files",
			"--enable-widec")
	})

	installBuildInstall("http://ftp.gnu.org/gnu/bash/bash-5.0.tar.gz", func() {
		DotConfigure("--prefix=/usr",
			"--docdir=/usr/share/doc/bash-5.0",
			"--without-bash-malloc",
			"--with-installed-readline")
		Make(NumberOfMakeJobs)
	}, func(destDir string) {
		Make("install", "DESTDIR="+destDir)
		Cd(destDir)
		Mkdir("bin")
		Mv("usr/bin/bash", "bin/bash")
		Ln("-s", "/bin/bash", "bin/sh")
	})

	installSimple("http://ftp.gnu.org/gnu/sed/sed-4.7.tar.xz")

	installConfigure("http://ftp.gnu.org/gnu/findutils/findutils-4.6.0.tar.gz", func() {
		Exec("bash", "-c", "sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' gl/lib/*.c")
		Exec("bash", "-c", "sed -i '/unistd/a #include <sys/sysmacros.h>' gl/lib/mountlist.c")
		appendToFile("gl/lib/stdio-impl.h", "#define _IO_IN_BACKUP 0x100")
		DotConfigure("--prefix=/usr", "--localstatedir=/var/lib/locate")
	})

	installSimple("http://www.greenwoodsoftware.com/less/less-551.tar.gz")
	installSimple("http://ftp.gnu.org/gnu/coreutils/coreutils-8.31.tar.xz")
	installSimple("https://github.com/vim/vim/archive/v8.1.1846/vim-8.1.1846.tar.gz")

	installConfigure("https://nchc.dl.sourceforge.net/project/procps-ng/Production/procps-ng-3.3.15.tar.xz", func() {
		DotConfigure("--prefix=/usr",
			"--exec-prefix=",
			"--libdir=/usr/lib",
			"--docdir=/usr/share/doc/procps-ng-3.3.15",
			"--disable-static",
			"--disable-kill")
	})

	installConfigure("https://www.kernel.org/pub/linux/utils/util-linux/v2.34/util-linux-2.34.tar.xz", func() {
		DotConfigure("--docdir=/usr/share/doc/util-linux-2.34",
			"--disable-chfn-chsh",
			"--disable-login",
			"--disable-nologin",
			"--disable-su",
			"--disable-wall",
			"--disable-setpriv",
			"--disable-runuser",
			"--disable-pylibmount",
			"--disable-static",
			"--without-python",
			"--without-systemd",
			"--without-systemdsystemunitdir",
			"--disable-makeinstall-chown",
		)
	})

	installSimple("http://ftp.gnu.org/gnu/gettext/gettext-0.20.1.tar.xz")
	installSimple("http://ftp.gnu.org/gnu/gawk/gawk-5.0.1.tar.xz")
	installSimple("http://ftp.gnu.org/gnu/bison/bison-3.5.tar.xz")
	installConfigure("http://ftp.gnu.org/gnu/make/make-4.2.1.tar.gz", func() {
		Exec("sh", "-c", "sed -i '211,217 d; 219,229 d; 232 d' glob/glob.c")
		DotConfigure("--prefix=/usr")
	})

	installConfigure("http://ftp.gnu.org/gnu/m4/m4-1.4.18.tar.xz", func() {
		Exec("sh", "-c", "sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' lib/*.c")
		Exec("sh", "-c", "echo \"#define _IO_IN_BACKUP 0x100\" >> lib/stdio-impl.h")
		DotConfigure("--prefix=/usr")
	})

	installSimple("http://ftp.gnu.org/gnu/gzip/gzip-1.10.tar.xz")

	// looks like we need this to bootstrap glibc
	installSimple("https://www.python.org/ftp/python/3.8.1/Python-3.8.1.tar.xz")

	installConfigure("https://github.com/shadow-maint/shadow/releases/download/4.8/shadow-4.8.tar.xz", func() {
		DotConfigure("--sysconfdir=/etc", "--with-group-name-max-length=32")
	})

	installConfigure("http://ftp.gnu.org/gnu/inetutils/inetutils-1.9.4.tar.xz", func() {
		DotConfigure("--prefix=/usr",
			"--localstatedir=/var",
			"--disable-logger",
			"--disable-whois",
			"--disable-rcp",
			"--disable-rexec",
			"--disable-rlogin",
			"--disable-rsh",
			"--disable-servers")
	})

	installConfigure("https://www.kernel.org/pub/linux/utils/net/iproute2/iproute2-5.4.0.tar.xz", func() {
		DotConfigure("--prefix=/usr")
	})

	installConfigure("https://roy.marples.name/downloads/dhcpcd/dhcpcd-8.1.4.tar.xz", func() {
		DotConfigure("--libexecdir=/lib/dhcpcd", "--dbdir=/var/lib/dhcpcd")
	})

	installBuildInstall(linuxKernelSourcesURL, func() {

		// zfs has very slow configure time, so disabling it until i get to zfs on root
		enableZFS := false

		if enableZFS {
			extract("https://github.com/zfsonlinux/zfs/releases/download/zfs-0.8.2/zfs-0.8.2.tar.gz")
			DotConfigure("--enable-linux-builtin")
			Make(NumberOfMakeJobs)
			Exec("./copy-builtin", "../linux-5.4.6")
			Cd("../linux-5.4.6")
			Rm("-rfv", "../zfs-0.8.2")
		}
		Make("defconfig")
		if enableZFS {
			appendToFile(".config", "CONFIG_ZFS=y")
		}
		appendToFile(".config",
			"CONFIG_CMDLINE_BOOL=y",
			"CONFIG_CMDLINE=\"rootwait root=/dev/sdc2 init=/sbin/minit\"",
			"CONFIG_DRM_NOUVEAU=y",
		)

		Make(NumberOfMakeJobs)
	}, func(destDir string) {
		//TODO dont forget modules as well
		Mkdir("-p", path.Join(destDir, "/boot/efi/EFI/boot"))
		Mv("arch/x86/boot/bzImage", path.Join(destDir, "/boot/efi/EFI/boot/bootx64.efi"))

		Make("headers")
		Mkdir("-p", path.Join(destDir, "usr"))
		Mv("usr/include", path.Join(destDir, "usr/"))

	})

	//TODO also package this so that we dont rebuild everything everytime
	Cd("minit")
	Exec("go", "build", "minit.go")
	Mv("minit", path.Join(InstallPrefix, "sbin/minit"))

	// Dev tools
	//	installConfigure("https://www.kernel.org/pub/software/scm/git/git-2.24.1.tar.xz", func() {
	//		dotConfigure("--prefix=/usr", "--without-tcltk")
	//	})

	installConfigure("https://www.openssl.org/source/openssl-1.1.1c.tar.gz", func() {
		Exec("./config", "--prefix=/usr",
			"--openssldir=/etc/ssl",
			"--libdir=lib",
			"shared",
			"zlib-dynamic")
	})

	installSimple("https://curl.haxx.se/download/curl-7.67.0.tar.xz")
	installSimple("http://ftp.gnu.org/gnu/tar/tar-1.32.tar.xz")
	installSimple("https://nchc.dl.sourceforge.net/project/lzmautils/xz-5.2.4.tar.xz")

	if *enterChroot {
		log.Printf("Entering chroot !!!!!!")
		Cd(Cwd)
		Exec("go", "build", "main.go")
		Exec("cp", "/etc/resolv.conf", path.Join(InstallPrefix, "etc"))
		Mv("main", path.Join(InstallPrefix, "root"))
		execCmdInteractive("sudo", "mount", "--bind", "/dev", path.Join(InstallPrefix, "dev"))
		execCmdInteractive("sudo", "chroot", InstallPrefix, "/root/main")
		os.Exit(1)
	}

}

// TODO extract directly onto usb stick
// TODO dont extract things in to newroot if they are already there
// TODO before extraction check for clashes
