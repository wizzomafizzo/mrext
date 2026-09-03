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
)

type app struct {
	name         string
	path         string
	bin          string
	releaseId    string
	reboot       bool
	inAll        bool
	releaseFiles []string
}

var apps = []app{
	{
		name: "contool",
		path: filepath.Join(cwd, "cmd", "contool"),
		bin:  "contool",
	},
	{
		name:      "remote",
		path:      filepath.Join(cwd, "cmd", "remote"),
		bin:       "remote.sh",
		releaseId: "mrext/remote",
		reboot:    true,
		inAll:     true,
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
		name: "samindex",
		path: filepath.Join(cwd, "cmd", "samindex"),
		bin:  "samindex",
	},
	{
		name:      "search",
		path:      filepath.Join(cwd, "cmd", "search"),
		bin:       "search.sh",
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
		name:      "playlog",
		path:      filepath.Join(cwd, "cmd", "playlog"),
		bin:       "playlog.sh",
		releaseId: "mrext/playlog",
		inAll:     true,
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
	_ = sh.Rm(generatedSystemMetadata)
}

func GenerateSystemMetadata() error {
	return sh.RunV("go", "run", "./internal/gensystemmetadata")
}

func buildApp(a app, out string, env map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	buildEnv := map[string]string{
		"CGO_ENABLED": "0",
		"GOPROXY":     "https://proxy.golang.org,direct",
	}
	for key, value := range env {
		buildEnv[key] = value
	}
	return sh.RunWithV(buildEnv, "go", "build", "-trimpath", "-o", out, a.path)
}

func buildApps(appName, platform string, env map[string]string) error {
	if appName == "all" {
		cleanPlatform(platform)
		for _, application := range apps {
			fmt.Println("Building", application.name)
			if err := buildApp(application, filepath.Join(binDir, platform, application.bin), env); err != nil {
				return err
			}
		}
		return nil
	}
	application := getApp(appName)
	if application == nil {
		return fmt.Errorf("unknown app: %s", appName)
	}
	return buildApp(*application, filepath.Join(binDir, platform, application.bin), env)
}

func Build(appName string) error {
	mg.Deps(GenerateSystemMetadata)
	platform := runtime.GOOS + "_" + runtime.GOARCH
	return buildApps(appName, platform, nil)
}

func Mister(appName string) error {
	mg.Deps(GenerateSystemMetadata)
	return buildApps(appName, "linux_arm", map[string]string{
		"GOOS":   "linux",
		"GOARCH": "arm",
		"GOARM":  "7",
	})
}

func UpdateExternalApps() {
	externalDir := filepath.Join(releasesDir, "external")
	_ = os.MkdirAll(externalDir, 0o755)
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

	if err := Mister(name); err != nil {
		fmt.Println("Error building MiSTer binary", err)
		os.Exit(1)
	}

	rd := filepath.Join(releasesDir, a.name)
	_ = os.MkdirAll(rd, 0o755)
	_ = os.MkdirAll(binReleasesDir, 0o755)
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
			err := os.Chmod(releaseBin, 0o755)
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
	_ = os.MkdirAll(binReleasesDir, 0o755)
	cleanPlatform("linux_arm")
	for _, app := range apps {
		if app.releaseId != "" {
			fmt.Println("Preparing release:", app.name)
			Release(app.name)
		}
	}
	UpdateExternalApps()
	for _, app := range externalApps {
		fmt.Println("Preparing release:", app.name)
		sh.Copy(filepath.Join(binReleasesDir, app.bin), filepath.Join(releasesDir, "external", app.bin))
	}
}

func Test() {
	mg.Deps(GenerateSystemMetadata)
	_ = sh.RunV("go", "test", "./...")
}

func Lint() error {
	mg.Deps(GenerateSystemMetadata)
	return sh.RunV("golangci-lint", "run", "./...")
}

func LintFix() error {
	mg.Deps(GenerateSystemMetadata)
	return sh.RunV("golangci-lint", "run", "--fix", "./...")
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
