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

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

var (
	cwd, _           = os.Getwd()
	binDir           = filepath.Join(cwd, "bin")
	releasesDir      = filepath.Join(cwd, "releases")
	releaseUrlPrefix = "https://github.com/wizzomafizzo/mrext/raw/main/releases"
	docsDir          = filepath.Join(cwd, "docs")
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
}

var apps = []app{
	{
		name:      "random",
		path:      filepath.Join(cwd, "cmd", "random"),
		bin:       "random.sh",
		releaseId: "mrext/random",
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
		ldFlags:   "-lcurses",
		releaseId: "mrext/search",
	},
	{
		name:      "launchsync",
		path:      filepath.Join(cwd, "cmd", "launchsync"),
		bin:       "launchsync.sh",
		releaseId: "mrext/launchsync",
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
	sh.Rm(filepath.Join(binDir, name))
}

func Clean() {
	sh.Rm(binDir)
	sh.Rm(armBuildCache)
	sh.Rm(armModCache)
	sh.Rm(kernelRepoPath)
}

func buildApp(a app, out string) {
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
	sh.RunV("sudo", "docker", "build", "--platform", "linux/arm/v7", "-t", armBuildImageName, armBuild)
}

func Mister(appName string) {
	buildCache := fmt.Sprintf("%s:%s", armBuildCache, "/home/build/.cache/go-build")
	os.Mkdir(armBuildCache, 0755)
	modCache := fmt.Sprintf("%s:%s", armModCache, "/home/build/go/pkg/mod")
	os.Mkdir(armModCache, 0755)
	buildDir := fmt.Sprintf("%s:%s", cwd, "/build")
	sh.RunV("sudo", "docker", "run", "--rm", "--platform", "linux/arm/v7", "-v", buildCache, "-v", modCache, "-v", buildDir, "--user", "1000:1000", armBuildImageName, "mage", "build", appName)
}

func UpdateExternalApps() {
	externalDir := filepath.Join(releasesDir, "external")
	os.MkdirAll(externalDir, 0755)
	for _, app := range externalApps {
		resp, err := http.Get(app.url)
		if err != nil || resp.StatusCode != 200 {
			fmt.Println("Error downloading", app.name, err)
			os.Exit(1)
		}
		defer resp.Body.Close()

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
	}
}

type updateDbFile struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
	Url  string `json:"url"`
}

type updateDbFolder struct {
	Tags []string `json:"tags"`
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
	defer file.Close()
	hash := md5.New()
	io.Copy(hash, file)
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func getFileSize(path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
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
		if app.releaseId == "" {
			continue
		}

		releaseBin := filepath.Join(releasesDir, app.name, app.bin)

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
			Url:  fmt.Sprintf("%s/%s/%s", releaseUrlPrefix, app.name, app.bin),
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
	dbFp.Write(dbFileJson)
	dbFp.Close()
}

func Release(name string) {
	a := getApp(name)
	if a == nil {
		fmt.Println("Unknown app", name)
		os.Exit(1)
	}
	platform := "linux_arm"
	mg.Deps(func() { cleanPlatform(platform) }, func() { Mister(name) })

	rd := filepath.Join(releasesDir, a.name)
	os.MkdirAll(rd, 0755)
	releaseBin := filepath.Join(rd, a.bin)
	err := sh.Copy(releaseBin, filepath.Join(binDir, platform, a.bin))
	if err != nil {
		fmt.Println("Error copying binary", err)
		os.Exit(1)
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
					Hash: hash,
					Size: size,
					Url:  fmt.Sprintf("%s/%s/%s", releaseUrlPrefix, a.name, a.bin),
				},
			},
			Folders: map[string]updateDbFolder{
				"Scripts": {},
			},
		}

		dbFileJson, _ := json.MarshalIndent(dbFile, "", "  ")
		dbFp, _ := os.Create(filepath.Join(rd, a.name+".json"))
		dbFp.Write(dbFileJson)
		dbFp.Close()
	}

	UpdateAllDb()
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

func GenerateSystemsDoc() {
	var systems []games.System
	for _, s := range games.Systems {
		systems = append(systems, s)
	}

	sort.Slice(systems, func(i, j int) bool {
		return systems[i].Id < systems[j].Id
	})

	md := "<!--- This file is automatically generated. Do not edit. --->\n\n"
	// TODO: some intro text and advice for requesting new systems
	md += "# Systems\n\n"
	md += "This is a list of all systems supported by the MiSTer Extensions scripts.\n\nPlease [open an issue](https://github.com/wizzomafizzo/mrext/issues/new) if you know a system that's not listed here but supports loading via MGL.\n\n"

	var toc []string
	for _, s := range systems {
		tocLink := "#" + strings.ReplaceAll(strings.ToLower(s.Name), " ", "-")
		toc = append(toc, fmt.Sprintf("[%s](%s)", s.Name, tocLink))
	}
	md += fmt.Sprintln(strings.Join(toc, ", "))

	md += "## Core Groups\n"
	md += "Core groups are aliases to multiple systems. They work as system IDs for all configuration options where a user must type a system ID manually. MiSTer Extensions differentiates between systems more than MiSTer itself, and these are included as a convenience so system folder names can still be used as IDs.\n\n"
	md += "| ID | Systems |\n| --- | --- |\n"
	cg := utils.MapKeys(games.CoreGroups)
	sort.Strings(cg)
	for _, k := range cg {
		syss := []string{}
		for _, s := range games.CoreGroups[k] {
			tocLink := "#" + strings.ReplaceAll(strings.ToLower(s.Name), " ", "-")
			syss = append(syss, fmt.Sprintf("[%s](%s)", s.Name, tocLink))
		}
		md += fmt.Sprintf("| %s | %s |\n", k, strings.Join(syss, ", "))
	}

	for _, s := range systems {
		md += fmt.Sprintln("\n##", s.Name)

		aliases := "-"
		if len(s.Alias) > 0 {
			aliases = strings.Join(s.Alias, ", ")
		}

		md += fmt.Sprintf("- **ID**: %s\n- **Aliases**: %s\n- **Folder**: %s\n- **RBF**: %s\n", s.Id, aliases, s.Folder, s.Rbf)

		if len(s.FileTypes) > 0 {
			md += "\n### Supported Files\n"
			md += fmt.Sprintf("\n| Files | Delay | Type | Index |\n| --- | --- | --- | --- |\n")

			for _, f := range s.FileTypes {
				files := "-"
				if len(f.Extensions) > 0 {
					files = strings.Join(f.Extensions, ", ")
				}

				delay := "-"
				fileType := "-"
				index := "-"

				if f.Mgl != nil {
					delay = fmt.Sprintf("%d", f.Mgl.Delay)
					fileType = f.Mgl.FileType
					index = fmt.Sprintf("%d", f.Mgl.Index)
				}

				md += fmt.Sprintf("| %s | %s | %s | %s |\n", files, delay, fileType, index)
			}
		}

		if len(s.AltRbf) > 0 {
			md += "\n### Alternate Cores\n"
			md += fmt.Sprintf("| Set | RBFs |\n| --- | --- |\n")
			for k, v := range s.AltRbf {
				md += fmt.Sprintf("| %s | %s |\n", k, strings.Join(v, ", "))
			}
		}
	}

	fp, _ := os.Create(filepath.Join(docsDir, "systems.md"))
	fp.WriteString(md)
	fp.Close()
}
