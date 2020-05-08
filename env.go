package kirill_linux

import (
	. "github.com/kirillrdy/kirill-linux/shell"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

type Env struct {
	// Cwd we remeber cwd in order to get back here
	Cwd string

	// DistfilesPath is where all distfiles are stored
	DistfilesPath string

	// BuildPath is where all the work is done
	BuildPath string

	// PkgPath is where binary packages get stored
	PkgPath string

	// InstallPrefix prefix is where we are all going to install things for new system
	InstallPrefix string
}

func (env Env) setUpGlobals() {

	var err error
	env.Cwd, err = os.Getwd()
	crash(err)

	env.DistfilesPath = path.Join(env.Cwd, "distfiles")
	err = os.MkdirAll(env.DistfilesPath, os.ModePerm)
	crash(err)

	env.BuildPath = path.Join(env.Cwd, "build")
	err = os.MkdirAll(env.BuildPath, os.ModePerm)
	crash(err)

	env.PkgPath = path.Join(env.Cwd, "package")
	err = os.MkdirAll(env.PkgPath, os.ModePerm)
	crash(err)

	env.InstallPrefix = path.Join(env.Cwd, "newroot")
	err = os.MkdirAll(env.InstallPrefix, os.ModePerm)
	crash(err)

}

func (env Env) fetch(url string) {
	log.Printf("fetching %s\n", url)
	Cd(env.DistfilesPath)

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

	Cd(env.Cwd)
}

func (env Env) extract(url string) {

	expetedTarPath := path.Join(env.DistfilesPath, path.Base(url))
	if _, err := os.Stat(expetedTarPath); os.IsNotExist(err) {
		env.fetch(url)
	}

	extractedSourcePath := path.Join(env.BuildPath, packageVersion(url))
	if _, err := os.Stat(extractedSourcePath); os.IsNotExist(err) {
		tarballPath := path.Join(env.DistfilesPath, path.Base(url))
		Tar("xf", tarballPath, "-C", env.BuildPath)
	}

	Cd(extractedSourcePath)
}

func (env Env) installSimple(url string) {
	env.installConfigure(url, func() {
		DotConfigure("--prefix=/usr")
	})
}

//TODO detect
const NumberOfMakeJobs = "-j8"

func (env Env) installConfigure(url string, configure func()) {
	env.installBuildInstall(url, func() {
		configure()

		Make(NumberOfMakeJobs)
	}, func(destDir string) {
		Make("install", "DESTDIR="+destDir)
	})
}

// stupid name, but what can you do
func (env Env) installBuildInstall(url string, build func(), install func(string)) {
	tarBall := path.Join(env.PkgPath, packageVersion(url)+".tar.xz")

	if _, err := os.Stat(tarBall); os.IsNotExist(err) {
		env.extract(url)
		sourceDir := path.Join(env.BuildPath, packageVersion(url))

		build()

		// Part of packaging
		destDir := path.Join(env.BuildPath, packageVersion(url)+"-package")
		install(destDir)
		Tar("cf", tarBall, "-C", destDir, ".")
		Cd(env.Cwd)
		Rm("-rf", destDir)
		Rm("-rf", sourceDir)
	}

	//TODO also dont do this if its already installed eg need some way of tracking those
	//TODO replace with desired prefix
	Tar("xf", tarBall, "-C", env.InstallPrefix)
}
