//go:build mage
// +build mage

package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/joho/godotenv/autoload"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const npmVersion = "12.0.2"

var (
	cwd, _                  = os.Getwd()
	binDir                  = filepath.Join(cwd, "_bin")
	binReleasesDir          = filepath.Join(binDir, "releases")
	releasesDir             = filepath.Join(cwd, "releases")
	releaseUrlPrefix        = "https://github.com/wizzomafizzo/mrext/releases/latest/download"
	generatedSystemMetadata = filepath.Join(cwd, "pkg", "games", "system_metadata.gen.json")
	remoteWebDir            = filepath.Join(cwd, "web", "remote")
	remoteWebBuildDir       = filepath.Join(cwd, "cmd", "remote", "_client", "build")
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
		name:      "favorites",
		path:      filepath.Join(cwd, "cmd", "favorites"),
		bin:       "favorites.sh",
		releaseId: "mrext/favorites",
		inAll:     true,
	},
	{
		name:      "bgm",
		path:      filepath.Join(cwd, "cmd", "bgm"),
		bin:       "bgm.sh",
		releaseId: "mrext/bgm",
		inAll:     true,
	},
	{
		name:      "gamesmenu",
		path:      filepath.Join(cwd, "cmd", "gamesmenu"),
		bin:       "gamesmenu.sh",
		releaseId: "mrext/gamesmenu",
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
	_ = sh.Rm(remoteWebBuildDir)
}

func RemoteWeb() error {
	npm := "npm@" + npmVersion
	if err := sh.RunV("npx", "--yes", npm, "--prefix", remoteWebDir, "ci"); err != nil {
		return fmt.Errorf("install Remote web dependencies: %w", err)
	}
	if err := sh.RunV("npx", "--yes", npm, "--prefix", remoteWebDir, "run", "build"); err != nil {
		return fmt.Errorf("build Remote web UI: %w", err)
	}
	return nil
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
	if appName == "remote" || appName == "all" {
		if err := RemoteWeb(); err != nil {
			return err
		}
	}
	platform := runtime.GOOS + "_" + runtime.GOARCH
	return buildApps(appName, platform, nil)
}

func Mister(appName string) error {
	mg.Deps(GenerateSystemMetadata)
	if appName == "remote" || appName == "all" {
		if err := RemoteWeb(); err != nil {
			return err
		}
	}
	return buildApps(appName, "linux_arm", map[string]string{
		"GOOS":   "linux",
		"GOARCH": "arm",
		"GOARM":  "7",
	})
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

	if err := Mister(name); err != nil {
		fmt.Println("Error building MiSTer binary", err)
		os.Exit(1)
	}

	if name == "remote" {
		clientIndex := filepath.Join(remoteWebBuildDir, "index.html")
		if _, err := os.Stat(clientIndex); err != nil {
			fmt.Println("Remote client build missing", clientIndex)
			os.Exit(1)
		}
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

// conflictMarkers are the markers git leaves in a file when a merge is not
// finished. One reached main once already, so this is checked, not trusted.
var conflictMarkers = []string{"<<<<<<< ", "=======", ">>>>>>> "}

// CheckConflicts fails when a tracked text file still contains merge conflict
// markers. Only "=======" paired with one of the other two counts, so tables
// and setext headings in Markdown do not trip it.
func CheckConflicts() error {
	tracked, err := sh.Output("git", "ls-files", "-z")
	if err != nil {
		return fmt.Errorf("list tracked files: %w", err)
	}

	var found []string
	for _, name := range strings.Split(tracked, "\x00") {
		if name == "" {
			continue
		}
		contents, readErr := os.ReadFile(name)
		if readErr != nil {
			// Unreadable or removed from the work tree; nothing to check.
			continue
		}
		if bytes.IndexByte(contents, 0) >= 0 {
			continue
		}
		var opened, separated bool
		for index, line := range strings.Split(string(contents), "\n") {
			switch {
			case strings.HasPrefix(line, conflictMarkers[0]):
				opened = true
			case opened && strings.TrimRight(line, "\r") == conflictMarkers[1]:
				separated = true
			case separated && strings.HasPrefix(line, conflictMarkers[2]):
				found = append(found, fmt.Sprintf("%s:%d", name, index+1))
				opened, separated = false, false
			}
		}
	}

	if len(found) > 0 {
		return fmt.Errorf("merge conflict markers in tracked files: %s", strings.Join(found, ", "))
	}
	return nil
}
