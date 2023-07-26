//go:build mage
// +build mage

package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

var (
	cwd, _           = os.Getwd()
	binDir           = filepath.Join(cwd, "_bin")
	binReleasesDir   = filepath.Join(binDir, "releases")
	releasesDir      = filepath.Join(cwd, "releases")
	releaseUrlPrefix = "https://github.com/wizzomafizzo/mrext/releases/latest/download"
	docsDir          = filepath.Join(cwd, "docs")
	upxBin           = os.Getenv("UPX_BIN")
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
	ldFlags   string
	releaseId string
	reboot    bool
	inAll     bool
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
	if runtime.GOOS != "linux" {
		_ = sh.RunV("docker", "build", "--platform", "linux/arm/v7", "-t", armBuildImageName, armBuild)
	} else {
		_ = sh.RunV("sudo", "docker", "build", "--platform", "linux/arm/v7", "-t", armBuildImageName, armBuild)
	}
}

func Mister(appName string) {
	buildCache := fmt.Sprintf("%s:%s", armBuildCache, "/home/build/.cache/go-build")
	_ = os.Mkdir(armBuildCache, 0755)
	modCache := fmt.Sprintf("%s:%s", armModCache, "/home/build/go/pkg/mod")
	_ = os.Mkdir(armModCache, 0755)
	buildDir := fmt.Sprintf("%s:%s", cwd, "/build")
	if runtime.GOOS != "linux" {
		_ = sh.RunV("docker", "run", "--rm", "--platform", "linux/arm/v7", "-v", buildCache, "-v", modCache, "-v", buildDir, "--user", "1000:1000", armBuildImageName, "mage", "build", appName)
	} else {
		_ = sh.RunV("sudo", "docker", "run", "--rm", "--platform", "linux/arm/v7", "-v", buildCache, "-v", modCache, "-v", buildDir, "--user", "1000:1000", armBuildImageName, "mage", "build", appName)
	}
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

func UpdateAllDb() {
	dbFile := updateDb{}
	dbFile.DbId = "mrext/all"
	dbFile.Timestamp = time.Now().Unix()
	dbFile.Files = make(map[string]updateDbFile)
	dbFile.Folders = map[string]updateDbFolder{
		"Scripts": {},
	}

	for _, app := range apps {
		if app.releaseId == "" || !app.inAll {
			continue
		}

		releaseBin := filepath.Join(binReleasesDir, app.bin)

		hash, err := getMd5Hash(releaseBin)
		if err != nil {
			fmt.Println("Error getting hash for", app.name, err)
			os.Exit(1)
		}

		size, err := getFileSize(releaseBin)
		if err != nil {
			fmt.Println("Error getting size for", app.name, err)
			os.Exit(1)
		}

		dbFile.Files["Scripts/"+app.bin] = updateDbFile{
			Hash:   hash,
			Size:   size,
			Url:    fmt.Sprintf("%s/%s", releaseUrlPrefix, app.bin),
			Reboot: app.reboot,
		}
	}

	for _, app := range externalApps {
		releaseBin := filepath.Join(releasesDir, "external", app.bin)

		hash, err := getMd5Hash(releaseBin)
		if err != nil {
			fmt.Println("Error getting hash for", app.name, err)
			os.Exit(1)
		}

		size, err := getFileSize(releaseBin)
		if err != nil {
			fmt.Println("Error getting size for", app.name, err)
			os.Exit(1)
		}

		dbFile.Files["Scripts/"+app.bin] = updateDbFile{
			Hash: hash,
			Size: size,
			Url:  app.url,
		}
	}

	dbFileJson, _ := json.MarshalIndent(dbFile, "", "  ")
	dbFp, _ := os.Create(filepath.Join(releasesDir, "all.json"))
	_, _ = dbFp.Write(dbFileJson)
	_ = dbFp.Close()
}

func Release(name string) {
	a := getApp(name)
	if a == nil {
		fmt.Println("Unknown app", name)
		os.Exit(1)
	}

	Mister(name)

	rd := filepath.Join(releasesDir, a.name)
	_ = os.MkdirAll(rd, 0755)
	releaseBin := filepath.Join(binReleasesDir, a.bin)
	err := sh.Copy(releaseBin, filepath.Join(binDir, "linux_arm", a.bin))
	if err != nil {
		fmt.Println("Error copying binary", err)
		os.Exit(1)
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

	if a.releaseId != "" {
		hash, err := getMd5Hash(releaseBin)
		if err != nil {
			fmt.Println("Error getting hash", a.name, err)
			os.Exit(1)
		}

		size, err := getFileSize(releaseBin)
		if err != nil {
			fmt.Println("Error getting size", a.name, err)
			os.Exit(1)
		}

		dbFile := updateDb{
			DbId:      a.releaseId,
			Timestamp: time.Now().Unix(),
			Files: map[string]updateDbFile{
				"Scripts/" + a.bin: {
					Hash:   hash,
					Size:   size,
					Url:    fmt.Sprintf("%s/%s", releaseUrlPrefix, a.bin),
					Reboot: a.reboot,
				},
			},
			Folders: map[string]updateDbFolder{
				"Scripts": {},
			},
		}

		dbFileJson, _ := json.MarshalIndent(dbFile, "", "  ")
		dbFp, _ := os.Create(filepath.Join(rd, a.name+".json"))
		_, _ = dbFp.Write(dbFileJson)
		_ = dbFp.Close()
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
	UpdateAllDb()
}

func MakeKernelImage() {
	_ = sh.RunV("sudo", "docker", "build", "-t", kernelBuildImageName, kernelBuild)
}

func Kernel() {
	if _, err := os.Stat(kernelRepoPath); os.IsNotExist(err) {
		_ = sh.RunV("git", "clone", "--depth", "1", kernelRepoUrl, kernelRepoPath)
	} else {
		_ = sh.RunV("git", "-C", kernelRepoPath, "reset", "--hard", "HEAD")
		_ = sh.RunV("git", "-C", kernelRepoPath, "clean", "-df")
		_ = sh.RunV("git", "-C", kernelRepoPath, "pull")
	}

	patches, _ := filepath.Glob(filepath.Join(kernelBuild, "*.patch"))
	for _, path := range patches {
		_ = sh.RunV("git", "-C", kernelRepoPath, "apply", path)
	}

	kCmd := sh.RunCmd("sudo", "docker", "run", "--rm", "-v", fmt.Sprintf("%s:%s", kernelRepoPath, "/build"), "--user", "1000:1000", kernelBuildImageName)
	_ = kCmd("make", "MiSTer_defconfig")
	_ = kCmd("make", "-j6", "zImage")
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

	if runtime.GOOS != "linux" {
		_ = sh.RunV("docker", "run", "--rm", "--platform", "linux/arm/v7", "-v", buildDir+":/build", "--user", "1000:1000", armBuildImageName, "bash", "./"+buildScript)
	} else {
		_ = sh.RunV("sudo", "docker", "run", "--rm", "--platform", "linux/arm/v7", "-v", buildDir+":/build", "--user", "1000:1000", armBuildImageName, "bash", "./"+buildScript)
	}
}

func Test() {
	_ = sh.RunV("go", "test", "./...")
}

func Coverage() {
	_ = sh.RunV("go", "test", "-coverprofile", "coverage.out", "./...")
	_ = sh.RunV("go", "tool", "cover", "-html", "coverage.out")
	_ = sh.Rm("coverage.out")
}

func GenSystemsDoc() {
	var systems []games.System
	for _, s := range games.Systems {
		systems = append(systems, s)
	}

	sort.Slice(systems, func(i, j int) bool {
		return systems[i].Name < systems[j].Name
	})

	md := "<!--- This file is automatically generated. Do not edit. --->\n\n"
	md += "# Systems\n\n"
	md += "This is a list of all systems supported by the MiSTer Extensions scripts. Please [open an issue](https://github.com/wizzomafizzo/mrext/issues/new) if a system is missing or not working.\n\n"

	var tocConsole []string
	var tocComputer []string
	var tocOther []string

	for _, s := range systems {
		tocAnchor := "#" + strings.ReplaceAll(strings.ToLower(s.Name), " ", "-")
		tocAnchor = utils.StripChars(tocAnchor, "()/")
		tocLink := fmt.Sprintf("[%s](%s)", s.Name, tocAnchor)

		if strings.HasPrefix(s.Rbf, "_Console") {
			tocConsole = append(tocConsole, tocLink)
		} else if strings.HasPrefix(s.Rbf, "_Computer") {
			tocComputer = append(tocComputer, tocLink)
		} else {
			tocOther = append(tocOther, tocLink)
		}
	}

	md += "**Consoles:** " + fmt.Sprintln(strings.Join(tocConsole, ", ")) + "\n\n"
	md += "**Computers:** " + fmt.Sprintln(strings.Join(tocComputer, ", ")) + "\n\n"
	md += "**Other:** " + fmt.Sprintln(strings.Join(tocOther, ", ")) + "\n\n"

	md += "## Core Groups\n"
	md += "Core groups are aliases to multiple systems. They work as system IDs for all configuration options where a user must type a system ID manually. MiSTer Extensions differentiates between systems more than MiSTer itself, and these are included as a convenience so system folder names can still be used as IDs.\n\n"
	md += "| ID | Systems |\n| --- | --- |\n"
	cg := utils.MapKeys(games.CoreGroups)
	sort.Strings(cg)
	for _, k := range cg {
		var syss []string
		for _, s := range games.CoreGroups[k] {
			tocLink := "#" + strings.ReplaceAll(strings.ToLower(s.Name), " ", "-")
			syss = append(syss, fmt.Sprintf("[%s](%s)", s.Name, tocLink))
		}
		md += fmt.Sprintf("| %s | %s |\n", k, strings.Join(syss, ", "))
	}

	for _, s := range systems {
		md += fmt.Sprintln("\n##", s.Name)

		var info []string

		info = append(info, fmt.Sprintf("**ID**: %s ", s.Id))

		if len(s.Alias) > 0 {
			aliases := strings.Join(s.Alias, ", ")
			info = append(info, fmt.Sprintf("**Aliases**: %s ", aliases))
		}

		info = append(info, fmt.Sprintf("**Folders**: %s", strings.Join(s.Folder, ", ")))
		info = append(info, fmt.Sprintf("**RBF**: %s", s.Rbf))

		md += "\n" + strings.Join(info, " | ") + "\n\n"

		if len(s.Slots) > 0 {
			md += fmt.Sprintf("\n| Label | Files | Delay | Type | Index |\n| --- | --- | --- | --- | --- |\n")

			for _, f := range s.Slots {
				files := "-"
				if len(f.Exts) > 0 {
					files = strings.Join(f.Exts, ", ")
				}

				label := "-"
				delay := "-"
				fileType := "-"
				index := "-"

				if f.Label != "" {
					label = f.Label
				}

				if f.Mgl != nil {
					delay = fmt.Sprintf("%d", f.Mgl.Delay)
					fileType = f.Mgl.Method
					index = fmt.Sprintf("%d", f.Mgl.Index)
				}

				md += fmt.Sprintf("| %s | %s | %s | %s | %s |\n", label, files, delay, fileType, index)
			}
		}

		if len(s.AltRbf) > 0 {
			md += "\n### Alternate Cores\n"
			md += fmt.Sprintf("| Set | RBFs |\n| --- | --- |\n")
			for k, v := range s.AltRbf {
				md += fmt.Sprintf("| %s | %s |\n", k, strings.Join(v, ", "))
			}
		}

		md += "\n[Back to top](#systems)\n"
	}

	fp, _ := os.Create(filepath.Join(docsDir, "systems.md"))
	_, _ = fp.WriteString(md)
	_ = fp.Close()
}
