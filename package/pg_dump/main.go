package main

import (
	"github.com/kirillrdy/kirill-linux"
	"github.com/kirillrdy/kirill-linux/config"
	"github.com/kirillrdy/kirill-linux/shell"
	"os"
)

func main() {
	config.Verbose = false
	env := kirill_linux.NewEnv()
	env.BuildInstall("https://ftp.postgresql.org/pub/source/v14.2/postgresql-14.2.tar.bz2", func() {
		shell.DotConfigure("--prefix="+env.InstallPrefix, "--with-uuid=e2fs")
		shell.Make(kirill_linux.NumberOfMakeJobs())
	}, func(destDir string) {
		shell.Make("install", "DESTDIR="+destDir)
		shell.Make("install", "DESTDIR="+destDir, "-C", "contrib")
	})

	shell.ExecInteractive("pg_dump", os.Args[1:]...)
}
