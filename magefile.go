//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	cwd, _           = os.Getwd()
	binDir           = filepath.Join(cwd, "bin")
	dockerBuild      = filepath.Join(cwd, "scripts", "docker")
	dockerImageName  = "mrext/mister"
	dockerBuildCache = filepath.Join(os.TempDir(), "mrext-buildcache")
	dockerModCache   = filepath.Join(os.TempDir(), "mrext-modcache")
)

type app struct {
	name      string
	path      string
	bin       string
	misterBin string
	ldFlags   string
}

var apps = []app{
	{
		name:    "search",
		path:    filepath.Join(cwd, "cmd", "search"),
		bin:     "search.sh",
		ldFlags: "-lcurses",
	},
	{
		name: "random",
		path: filepath.Join(cwd, "cmd", "random"),
		bin:  "random.sh",
	},
	{
		name: "samindex",
		path: filepath.Join(cwd, "cmd", "samindex"),
		bin:  "samindex",
	},
}

var Default = Build

func cleanPlatform(name string) {
	sh.Rm(filepath.Join(binDir, name))
}

func Clean() {
	sh.Rm(binDir)
	sh.Rm(dockerBuildCache)
	sh.Rm(dockerModCache)
}

func buildApp(a app, out string) {
	fmt.Println("Building", a.name)
	if a.ldFlags == "" {
		sh.RunV("go", "build", "-o", out, a.path)
	} else {
		staticEnv := map[string]string{
			"CGO_ENABLED": "1",
			"CGO_LDFLAGS": a.ldFlags,
		}
		sh.RunWithV(staticEnv, "go", "build", "--ldflags", "-linkmode external -extldflags -static", "-o", out, a.path)
	}
}

func Build() {
	platform := runtime.GOOS + "_" + runtime.GOARCH
	mg.Deps(func() { cleanPlatform(platform) })
	for _, app := range apps {
		buildApp(app, filepath.Join(binDir, platform, app.bin))
	}
}

func BuildDockerImage() {
	sh.RunV("sudo", "docker", "build", "--platform", "linux/arm/v7", "-t", dockerImageName, dockerBuild)
}

func Mister() {
	buildCache := fmt.Sprintf("%s:%s", dockerBuildCache, "/home/build/.cache/go-build")
	os.Mkdir(dockerBuildCache, 0755)
	modCache := fmt.Sprintf("%s:%s", dockerModCache, "/home/build/go/pkg/mod")
	os.Mkdir(dockerModCache, 0755)
	buildDir := fmt.Sprintf("%s:%s", cwd, "/build")
	sh.RunV("sudo", "docker", "run", "--rm", "--platform", "linux/arm/v7", "-v", buildCache, "-v", modCache, "-v", buildDir, "--user", "1000:1000", dockerImageName, "mage", "build")
}
