package mister

import (
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

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

	err = xml.Unmarshal(file, &mgl)
	if err != nil {
		return mgl, err
	}

	return mgl, nil
}
