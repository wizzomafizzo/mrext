package mister

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func GetActiveCoreName() (string, error) {
	data, err := os.ReadFile(config.CoreNameFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func ActiveGameEnabled() bool {
	_, err := os.Stat(config.ActiveGameFile)
	return err == nil
}

func SetActiveGame(path string) error {
	file, err := os.Create(config.ActiveGameFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(path)
	if err != nil {
		return err
	}

	return nil
}

func GetActiveGame() (string, error) {
	data, err := os.ReadFile(config.ActiveGameFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Convert a launchable path to an absolute path.
func ResolvePath(path string) string {
	if path == "" {
		return path
	}

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(config.SdFolder)

	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	return abs
}

// Search for directories in root that start with "_".
func GetMenuFolders(root string) []string {
	var folders []string

	// TODO: confirm menu can't traverse symlinks
	var scan func(path string)
	scan = func(folder string) {
		files, err := os.ReadDir(folder)
		if err != nil {
			return
		}
		for _, file := range files {
			if file.IsDir() && file.Name()[0] == '_' {
				path := filepath.Join(folder, file.Name())
				folders = append(folders, path)
				scan(path)
			}
		}
	}

	scan(root)
	return folders
}

func isRbf(path string) bool {
	return filepath.Ext(strings.ToLower(path)) == ".rbf"
}

// Do a shallow search of RBF files in root and return list of relative paths.
func GetRbfs(root string) []string {
	var rbfs []string

	rootFiles, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	for _, rootFile := range rootFiles {
		if !rootFile.IsDir() && isRbf(rootFile.Name()) {
			rbfs = append(rbfs, rootFile.Name())
		} else if rootFile.IsDir() && rootFile.Name()[0] == '_' {
			subFiles, err := os.ReadDir(filepath.Join(root, rootFile.Name()))
			if err != nil {
				continue
			}

			for _, subFile := range subFiles {
				if !subFile.IsDir() && isRbf(subFile.Name()) {
					rbfs = append(rbfs, rootFile.Name()+"/"+subFile.Name())
				}
			}
		}
	}

	return rbfs
}

// Find an RBF in a list of all RBFs and return a value suitable for MGL.
func MatchRbf(rbfs []string, match string) string {
	if len(rbfs) == 0 {
		return ""
	}

	for _, rbf := range rbfs {
		parts := strings.Split(rbf, "/")
		file := parts[len(parts)-1]
		if strings.HasPrefix(strings.ToLower(file), strings.ToLower(match)) {
			if len(parts) == 1 {
				return match
			} else {
				return strings.Join(append(parts[0:len(parts)-1], match), "/")
			}
		}
	}

	return ""
}

type RecentEntry struct {
	Directory string
	Name      string
	Label     string
}

func ReadRecent(path string) ([]RecentEntry, error) {
	var recents []RecentEntry

	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	for {
		entry := make([]byte, 1024+256+256)
		n, err := file.Read(entry)
		if err == io.EOF || n == 0 {
			break
		} else if err != nil {
			return nil, err
		}

		empty := true
		for _, b := range entry {
			if b != 0 {
				empty = false
			}
		}
		if empty {
			break
		}

		recents = append(recents, RecentEntry{
			Directory: strings.Trim(string(entry[:1024]), "\x00"),
			Name:      strings.Trim(string(entry[1024:1280]), "\x00"),
			Label:     strings.Trim(string(entry[1280:1536]), "\x00"),
		})
	}

	return recents, nil
}

type MGLFile struct {
	XMLName xml.Name `xml:"file"`
	Delay   int      `xml:"delay,attr"`
	Type    string   `xml:"type,attr"`
	Index   int      `xml:"index,attr"`
	Path    string   `xml:"path,attr"`
}

type MGL struct {
	XMLName xml.Name `xml:"mistergamedescription"`
	Rbf     string   `xml:"rbf"`
	SetName string   `xml:"setname"`
	File    MGLFile  `xml:"file"`
}

func ReadMgl(path string) (MGL, error) {
	var mgl MGL

	if _, err := os.Stat(path); err != nil {
		return mgl, err
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return mgl, err
	}

	decoder := xml.NewDecoder(bytes.NewReader(file))
	decoder.Strict = false

	err = decoder.Decode(&mgl)
	if err != nil {
		return mgl, err
	}

	return mgl, nil
}

type MenuConfig struct {
	BackgroundMode int
}

const (
	BackgroundModeNone      = 0
	BackgroundModeWallpaper = 2
	BackgroundModeHBars1    = 4
	BackgroundModeHBars2    = 6
	BackgroundModeVBars1    = 8
	BackgroundModeVBars2    = 10
	BackgroundModeSpectrum  = 12
	BackgroundModeBlack     = 14
)

func ReadMenuConfig() (MenuConfig, error) {
	var cfg MenuConfig

	if _, err := os.Stat(config.MenuConfigFile); err != nil {
		return cfg, err
	}

	file, err := os.ReadFile(config.MenuConfigFile)
	if err != nil {
		return cfg, err
	}

	cfg.BackgroundMode = int(file[0])

	return cfg, nil
}

func SetMenuBackgroundMode(mode int) error {
	if !utils.Contains([]int{
		BackgroundModeNone,
		BackgroundModeWallpaper,
		BackgroundModeHBars1,
		BackgroundModeHBars2,
		BackgroundModeVBars1,
		BackgroundModeVBars2,
		BackgroundModeSpectrum,
		BackgroundModeBlack,
	}, mode) {
		return fmt.Errorf("invalid background mode")
	}

	cfg, err := ReadMenuConfig()
	if err != nil {
		return err
	}

	if cfg.BackgroundMode == mode {
		return nil
	}

	file, err := os.ReadFile(config.MenuConfigFile)
	if err != nil {
		return err
	}

	file[0] = byte(mode)

	return os.WriteFile(config.MenuConfigFile, file, 0644)
}

func GetMounts(cfg *config.UserConfig) ([]string, error) {
	file, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil, err
	}

	var mounts []string
	gamesFolders := games.GetGamesFolders(cfg)

	for _, line := range strings.Split(string(file), "\n") {
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")

		if len(parts) < 2 {
			continue
		}

		if utils.Contains(gamesFolders, parts[1]) {
			mounts = append(mounts, parts[1])
		}
	}

	return mounts, nil
}

type DiskUsage struct {
	Total uint64
	Free  uint64
	Used  uint64
}

func GetDiskUsage(path string) (DiskUsage, error) {
	var usage DiskUsage

	stat := syscall.Statfs_t{}
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return usage, err
	}

	usage.Total = stat.Blocks * uint64(stat.Bsize)
	usage.Free = stat.Bfree * uint64(stat.Bsize)
	usage.Used = usage.Total - usage.Free

	return usage, nil
}
