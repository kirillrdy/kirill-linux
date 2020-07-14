package main

import (
	"github.com/kirillrdy/kirill-linux"
	"github.com/kirillrdy/kirill-linux/shell"
	"os"
)

func main() {
	env := kirill_linux.NewEnv()
	env.BuildInstall("https://ftp.postgresql.org/pub/source/v12.3/postgresql-12.3.tar.bz2", func() {
		shell.DotConfigure("--prefix="+env.InstallPrefix, "--with-uuid=e2fs")
		shell.Make(kirill_linux.NumberOfMakeJobs())
	}, func(destDir string) {
		shell.Make("install", "DESTDIR="+destDir)
		shell.Make("install", "DESTDIR="+destDir, "-C", "contrib")
	})

	shell.ExecInteractive("pg_dump", os.Args[1:]...)
}
