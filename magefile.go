//go:build mage
// +build mage

package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/joho/godotenv/autoload"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	cwd, _                  = os.Getwd()
	binDir                  = filepath.Join(cwd, "_bin")
	binReleasesDir          = filepath.Join(binDir, "releases")
	releasesDir             = filepath.Join(cwd, "releases")
	releaseUrlPrefix        = "https://github.com/wizzomafizzo/mrext/releases/latest/download"
	generatedSystemMetadata = filepath.Join(cwd, "pkg", "games", "system_metadata.gen.json")
	upxBin                  = os.Getenv("UPX_BIN")
	// docker arm build
	armBuild          = filepath.Join(cwd, "scripts", "armbuild")
	armBuildImageName = "mrext/armbuild"
	armBuildCache     = filepath.Join(os.TempDir(), "mrext-buildcache")
	armModCache       = filepath.Join(os.TempDir(), "mrext-modcache")
	// docker kernel build
	kernelBuild          = filepath.Join(cwd, "scripts", "kernelbuild")
	kernelBuildImageName = "mrext/kernelbuild"
	kernelRepoName       = "Linux-Kernel_MiSTer"
	kernelRepoPath       = filepath.Join(kernelBuild, "_build", kernelRepoName)
	kernelRepoUrl        = fmt.Sprintf("https://github.com/MiSTer-devel/%s.git", kernelRepoName)
)

type app struct {
	name         string
	path         string
	bin          string
	ldFlags      string
	releaseId    string
	reboot       bool
	inAll        bool
	releaseFiles []string
}

var apps = []app{
	{
		name: "background",
		path: filepath.Join(cwd, "cmd", "background"),
		bin:  "background",
	},
	{
		name: "contool",
		path: filepath.Join(cwd, "cmd", "contool"),
		bin:  "contool",
	},
	{
		name:      "remote",
		path:      filepath.Join(cwd, "cmd", "remote"),
		bin:       "remote.sh",
		ldFlags:   "-lcurses",
		releaseId: "mrext/remote",
		reboot:    true,
		inAll:     true,
	},
	{
		name: "favorites",
		path: filepath.Join(cwd, "cmd", "favorites"),
		bin:  "addfav",
		// releaseId: "mrext/favorites",
	},
	{
		name:      "lastplayed",
		path:      filepath.Join(cwd, "cmd", "lastplayed"),
		bin:       "lastplayed.sh",
		releaseId: "mrext/lastplayed",
		inAll:     true,
	},
	{
		name:      "random",
		path:      filepath.Join(cwd, "cmd", "random"),
		bin:       "random.sh",
		releaseId: "mrext/random",
		inAll:     true,
	},
	{
		name:         "nfc",
		path:         filepath.Join(cwd, "cmd", "nfc"),
		bin:          "nfc.sh",
		ldFlags:      "-lnfc -lusb -lcurses",
		releaseFiles: []string{filepath.Join(cwd, "scripts", "nfcui", "nfcui.sh")},
	},
	{
		name: "samindex",
		path: filepath.Join(cwd, "cmd", "samindex"),
		bin:  "samindex",
	},
	{
		name: "screenshots",
		path: filepath.Join(cwd, "cmd", "screenshots"),
		bin:  "screenshots.sh",
		// releaseId: "mrext/screenshots",
	},
	{
		name:      "search",
		path:      filepath.Join(cwd, "cmd", "search"),
		bin:       "search.sh",
		ldFlags:   "-lcurses",
		releaseId: "mrext/search",
		inAll:     true,
	},
	{
		name:      "launchsync",
		path:      filepath.Join(cwd, "cmd", "launchsync"),
		bin:       "launchsync.sh",
		releaseId: "mrext/launchsync",
		inAll:     true,
	},
	{
		name:      "launchseq",
		path:      filepath.Join(cwd, "cmd", "launchseq"),
		bin:       "launchseq.sh",
		releaseId: "mrext/launchseq",
	},
	{
		name:      "playlog",
		path:      filepath.Join(cwd, "cmd", "playlog"),
		bin:       "playlog.sh",
		releaseId: "mrext/playlog",
		inAll:     true,
	},
	{
		name: "vplay",
		path: filepath.Join(cwd, "cmd", "vplay"),
		bin:  "vplay.sh",
	},
	{
		name: "mm",
		path: filepath.Join(cwd, "cmd", "mm"),
		bin:  "mm",
	},
}

type externalApp struct {
	name string
	url  string
	bin  string
}

var externalApps = []externalApp{
	{
		name: "bgm",
		url:  "https://github.com/wizzomafizzo/MiSTer_BGM/raw/main/bgm.sh",
		bin:  "bgm.sh",
	},
	{
		name: "favorites",
		url:  "https://github.com/wizzomafizzo/MiSTer_Favorites/raw/main/favorites.sh",
		bin:  "favorites.sh",
	},
	{
		name: "gamesmenu",
		url:  "https://github.com/wizzomafizzo/MiSTer_GamesMenu/raw/main/gamesmenu.sh",
		bin:  "gamesmenu.sh",
	},
}

var scriptApps = []externalApp{
	{
		name: "pocketbackup",
		bin:  "pocketbackup.sh",
	},
}

func getApp(name string) *app {
	for _, a := range apps {
		if a.name == name {
			return &a
		}
	}
	return nil
}

func cleanPlatform(name string) {
	_ = sh.Rm(filepath.Join(binDir, name))
}

func Clean() {
	_ = sh.Rm(binDir)
	_ = sh.Rm(armBuildCache)
	_ = sh.Rm(armModCache)
	_ = sh.Rm(kernelRepoPath)
	_ = sh.Rm(generatedSystemMetadata)
}

func GenerateSystemMetadata() error {
	return sh.RunV("go", "run", "./internal/gensystemmetadata")
}

func buildApp(a app, out string) {
	if a.ldFlags == "" {
		env := map[string]string{
			"GOPROXY": "https://goproxy.io,direct",
		}
		_ = sh.RunWithV(env, "go", "build", "-o", out, a.path)
	} else {
		staticEnv := map[string]string{
			"GOPROXY":     "https://goproxy.io,direct",
			"CGO_ENABLED": "1",
			"CGO_LDFLAGS": a.ldFlags,
		}
		_ = sh.RunWithV(staticEnv, "go", "build", "--ldflags", "-linkmode external -extldflags -static", "-o", out, a.path)
	}
}

func Build(appName string) {
	mg.Deps(GenerateSystemMetadata)
	platform := runtime.GOOS + "_" + runtime.GOARCH
	if appName == "all" {
		mg.Deps(func() { cleanPlatform(platform) })
		for _, app := range apps {
			fmt.Println("Building", app.name)
			buildApp(app, filepath.Join(binDir, platform, app.bin))
		}
	} else {
		app := getApp(appName)
		if app == nil {
			fmt.Println("Unknown app", appName)
			os.Exit(1)
		}
		buildApp(*app, filepath.Join(binDir, platform, app.bin))
	}
}

func MakeArmImage() {
	_ = sh.RunV("docker", "build", "--platform", "linux/arm/v7", "-t", armBuildImageName, armBuild)
}

func Mister(appName string) {
	mg.Deps(GenerateSystemMetadata)
	buildCache := fmt.Sprintf("%s:%s", armBuildCache, "/home/build/.cache/go-build")
	_ = os.Mkdir(armBuildCache, 0755)
	modCache := fmt.Sprintf("%s:%s", armModCache, "/home/build/go/pkg/mod")
	_ = os.Mkdir(armModCache, 0755)
	buildDir := fmt.Sprintf("%s:%s", cwd, "/build")
	_ = sh.RunV("docker", "run", "--rm", "--platform", "linux/arm/v7", "-v", buildCache, "-v", modCache, "-v", buildDir, "--user", "1000:1000", armBuildImageName, "mage", "build", appName)
}

func UpdateExternalApps() {
	externalDir := filepath.Join(releasesDir, "external")
	_ = os.MkdirAll(externalDir, 0755)
	for _, app := range externalApps {
		resp, err := http.Get(app.url)
		if err != nil || resp.StatusCode != 200 {
			fmt.Println("Error downloading", app.name, err)
			os.Exit(1)
		}

		out, err := os.Create(filepath.Join(externalDir, app.bin))
		if err != nil {
			fmt.Println("Error creating", app.name, err)
			os.Exit(1)
		}

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			fmt.Println("Error writing", app.name, err)
			os.Exit(1)
		}

		_ = resp.Body.Close()
	}
}

type updateDbFile struct {
	Hash   string `json:"hash"`
	Size   int64  `json:"size"`
	Url    string `json:"url"`
	Reboot bool   `json:"reboot,omitempty"`
}

type updateDbFolder struct {
	Tags []string `json:"tags,omitempty"`
}

type updateDb struct {
	DbId      string                    `json:"db_id"`
	Timestamp int64                     `json:"timestamp"`
	Files     map[string]updateDbFile   `json:"files"`
	Folders   map[string]updateDbFolder `json:"folders"`
}

func getMd5Hash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	_, _ = io.Copy(hash, file)
	_ = file.Close()
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func getFileSize(path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return 0, err
	}

	size := stat.Size()
	_ = file.Close()

	return size, nil
}

func Release(name string) {
	a := getApp(name)
	if a == nil {
		fmt.Println("Unknown app", name)
		os.Exit(1)
	}

	if name == "remote" {
		clientIndex := filepath.Join(cwd, "cmd", "remote", "_client", "build", "index.html")
		if _, err := os.Stat(clientIndex); err != nil {
			fmt.Println("Remote client build missing", clientIndex)
			os.Exit(1)
		}
	}

	if runtime.GOOS == "linux" && runtime.GOARCH == "arm" {
		Build(name)
	} else {
		Mister(name)
	}

	rd := filepath.Join(releasesDir, a.name)
	_ = os.MkdirAll(rd, 0755)
	_ = os.MkdirAll(binReleasesDir, 0755)
	releaseBin := filepath.Join(binReleasesDir, a.bin)
	err := sh.Copy(releaseBin, filepath.Join(binDir, "linux_arm", a.bin))
	if err != nil {
		fmt.Println("Error copying binary", err)
		os.Exit(1)
	}

	for _, f := range a.releaseFiles {
		err := sh.Copy(filepath.Join(binReleasesDir, filepath.Base(f)), f)
		if err != nil {
			fmt.Println("Error copying release file", err)
			os.Exit(1)
		}
	}

	if upxBin == "" {
		fmt.Println("UPX is required for releases")
		os.Exit(1)
	} else {
		if runtime.GOOS != "windows" {
			err := os.Chmod(releaseBin, 0755)
			if err != nil {
				fmt.Println("Error chmod release bin", err)
				os.Exit(1)
			}
		}

		err := sh.RunV(upxBin, "-9", releaseBin)
		if err != nil {
			fmt.Println("Error compressing binary", err)
			os.Exit(1)
		}
	}
}

func PrepRelease() {
	_ = sh.Rm(binReleasesDir)
	_ = os.MkdirAll(binReleasesDir, 0755)
	cleanPlatform("linux_arm")
	for _, app := range apps {
		if app.releaseId != "" {
			fmt.Println("Preparing release:", app.name)
			Release(app.name)
		}
	}
	for _, app := range scriptApps {
		fmt.Println("Preparing release:", app.name)
		sh.Copy(filepath.Join(binReleasesDir, app.bin), filepath.Join(cwd, "scripts", app.name, app.bin))
	}
	UpdateExternalApps()
	for _, app := range externalApps {
		fmt.Println("Preparing release:", app.name)
		sh.Copy(filepath.Join(binReleasesDir, app.bin), filepath.Join(releasesDir, "external", app.bin))
	}
}

func MakeKernelImage() {
	_ = sh.RunV("docker", "build", "-t", kernelBuildImageName, kernelBuild)
}

func Kernel() {
	if _, err := os.Stat(kernelRepoPath); os.IsNotExist(err) {
		_ = sh.RunV("git", "clone", "--depth", "1", kernelRepoUrl, kernelRepoPath)
	}

	patches, _ := filepath.Glob(filepath.Join(kernelBuild, "*.patch"))
	for _, path := range patches {
		_ = sh.RunV("git", "-C", kernelRepoPath, "apply", path)
	}

	kCmd := sh.RunCmd("docker", "run", "--rm", "-v", fmt.Sprintf("%s:%s", kernelRepoPath, "/build"), "--user", "1000:1000", kernelBuildImageName)
	_ = kCmd("make", "MiSTer_defconfig")
	_ = kCmd("make", "modules")
	_ = kCmd("make", "-j16", "zImage")
	_ = kCmd("make", "socfpga_cyclone5_de10_nano.dtb")

	zImage, _ := os.Open(filepath.Join(kernelRepoPath, "arch", "arm", "boot", "zImage"))
	dtb, _ := os.Open(filepath.Join(kernelRepoPath, "arch", "arm", "boot", "dts", "socfpga_cyclone5_de10_nano.dtb"))

	_ = os.MkdirAll(filepath.Join(binDir, "linux"), 0755)
	kernel, _ := os.Create(filepath.Join(binDir, "linux", "zImage_dtb"))

	_, _ = io.Copy(kernel, zImage)
	_, _ = io.Copy(kernel, dtb)

	_ = kernel.Close()
	_ = dtb.Close()
	_ = zImage.Close()
}

func MakeArmApp(name string) {
	buildScript := name + ".sh"
	if _, err := os.Stat(filepath.Join(armBuild, buildScript)); os.IsNotExist(err) {
		fmt.Println("No build script for", name)
		os.Exit(1)
	}

	buildDir := filepath.Join(armBuild, "_build")
	_ = os.MkdirAll(buildDir, 0755)

	err := sh.Copy(filepath.Join(buildDir, buildScript), filepath.Join(armBuild, buildScript))
	if err != nil {
		fmt.Println("Error copying build script", err)
		os.Exit(1)
	}

	_ = sh.RunV("docker", "run", "--rm", "--platform", "linux/arm/v7", "-v", buildDir+":/build", "--user", "1000:1000", armBuildImageName, "bash", "./"+buildScript)
}

func Test() {
	mg.Deps(GenerateSystemMetadata)
	_ = sh.RunV("go", "test", "./...")
}

func Coverage() {
	mg.Deps(GenerateSystemMetadata)
	_ = sh.RunV("go", "test", "-coverprofile", "coverage.out", "./...")
	_ = sh.RunV("go", "tool", "cover", "-html", "coverage.out")
	_ = sh.Rm("coverage.out")
}

func GenSystemsDoc() {
	mg.Deps(GenerateSystemMetadata)
	_ = sh.RunV("go", "run", "./internal/gensystemsdoc")
}
