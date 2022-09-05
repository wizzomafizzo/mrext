//go:build mage
// +build mage

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	cwd, _ = os.Getwd()
	binDir = filepath.Join(cwd, "bin")
	// docker arm build
	armBuild          = filepath.Join(cwd, "scripts", "armbuild")
	armBuildImageName = "mrext/armbuild"
	armBuildCache     = filepath.Join(os.TempDir(), "mrext-buildcache")
	armModCache       = filepath.Join(os.TempDir(), "mrext-modcache")
	// docker kernel build
	kernelBuild          = filepath.Join(cwd, "scripts", "kernelbuild")
	kernelBuildImageName = "mrext/kernelbuild"
	kernelRepoName       = "Linux-Kernel_MiSTer"
	kernelRepoPath       = filepath.Join(os.TempDir(), "mrext-"+kernelRepoName)
	kernelRepoUrl        = fmt.Sprintf("https://github.com/MiSTer-devel/%s.git", kernelRepoName)
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
		name: "random",
		path: filepath.Join(cwd, "cmd", "random"),
		bin:  "random.sh",
	},
	{
		name: "samindex",
		path: filepath.Join(cwd, "cmd", "samindex"),
		bin:  "samindex",
	},
	{
		name:    "search",
		path:    filepath.Join(cwd, "cmd", "search"),
		bin:     "search.sh",
		ldFlags: "-lcurses",
	},
	{
		name: "launchsync",
		path: filepath.Join(cwd, "cmd", "launchsync"),
		bin:  "launchsync.sh",
	},
}

var Default = Build

func cleanPlatform(name string) {
	sh.Rm(filepath.Join(binDir, name))
}

func Clean() {
	sh.Rm(binDir)
	sh.Rm(armBuildCache)
	sh.Rm(armModCache)
	sh.Rm(kernelRepoPath)
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

func MakeArmImage() {
	sh.RunV("sudo", "docker", "build", "--platform", "linux/arm/v7", "-t", armBuildImageName, armBuild)
}

func Mister() {
	buildCache := fmt.Sprintf("%s:%s", armBuildCache, "/home/build/.cache/go-build")
	os.Mkdir(armBuildCache, 0755)
	modCache := fmt.Sprintf("%s:%s", armModCache, "/home/build/go/pkg/mod")
	os.Mkdir(armModCache, 0755)
	buildDir := fmt.Sprintf("%s:%s", cwd, "/build")
	sh.RunV("sudo", "docker", "run", "--rm", "--platform", "linux/arm/v7", "-v", buildCache, "-v", modCache, "-v", buildDir, "--user", "1000:1000", armBuildImageName, "mage", "build")
}

func MakeKernelImage() {
	sh.RunV("sudo", "docker", "build", "-t", kernelBuildImageName, kernelBuild)
}

func Kernel() {
	if _, err := os.Stat(kernelRepoPath); os.IsNotExist(err) {
		sh.RunV("git", "clone", "--depth", "1", kernelRepoUrl, kernelRepoPath)
	} else {
		sh.RunV("git", "-C", kernelRepoPath, "reset", "--hard", "HEAD")
		sh.RunV("git", "-C", kernelRepoPath, "clean", "-df")
		sh.RunV("git", "-C", kernelRepoPath, "pull")
	}

	patches, _ := filepath.Glob(filepath.Join(kernelBuild, "*.patch"))
	for _, path := range patches {
		sh.RunV("git", "-C", kernelRepoPath, "apply", path)
	}

	kCmd := sh.RunCmd("sudo", "docker", "run", "--rm", "-v", fmt.Sprintf("%s:%s", kernelRepoPath, "/build"), "--user", "1000:1000", kernelBuildImageName)
	kCmd("make", "MiSTer_defconfig")
	kCmd("make", "-j6", "zImage")
	kCmd("make", "socfpga_cyclone5_de10_nano.dtb")

	zImage, _ := os.Open(filepath.Join(kernelRepoPath, "arch", "arm", "boot", "zImage"))
	defer zImage.Close()
	dtb, _ := os.Open(filepath.Join(kernelRepoPath, "arch", "arm", "boot", "dts", "socfpga_cyclone5_de10_nano.dtb"))
	defer dtb.Close()

	os.MkdirAll(filepath.Join(binDir, "linux"), 0755)
	kernel, _ := os.Create(filepath.Join(binDir, "linux", "zImage_dtb"))
	defer kernel.Close()

	io.Copy(kernel, zImage)
	io.Copy(kernel, dtb)
}

func Test() {
	sh.RunV("go", "test", "./...")
}

func Coverage() {
	sh.RunV("go", "test", "-coverprofile", "coverage.out", "./...")
	sh.RunV("go", "tool", "cover", "-html", "coverage.out")
	sh.Rm("coverage.out")
}
