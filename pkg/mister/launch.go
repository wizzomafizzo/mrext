package mister

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	s "strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
)

func GenerateMgl(system *games.System, path string) (string, error) {
	var mglDef *games.MglParams

	for _, ft := range system.Slots {
		for _, ext := range ft.Exts {
			if s.HasSuffix(s.ToLower(path), ext) {
				mglDef = ft.Mgl
			}
		}
	}

	if mglDef == nil {
		return "", fmt.Errorf("system has no matching mgl args: %s, %s", system.Id, path)
	} else {
		// TODO: generate this from xml
		return fmt.Sprintf(
			"<mistergamedescription>\n\t<rbf>%s</rbf>\n\t<file delay=\"%d\" type=\"%s\" index=\"%d\" path=\"../../../../..%s\"/>\n</mistergamedescription>\n",
			system.Rbf, mglDef.Delay, mglDef.Method, mglDef.Index, path,
		), nil
	}
}

// TODO: move to utils?
func writeTempFile(content string, ext string) (string, error) {
	tmpFile, err := ioutil.TempFile("", "*."+ext)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(content)
	if err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func launchFile(path string) error {
	_, err := os.Stat(config.CmdInterface)
	if err != nil {
		return fmt.Errorf("command interface not accessible: %s", err)
	}

	if !(s.HasSuffix(s.ToLower(path), ".mgl") || s.HasSuffix(s.ToLower(path), ".mra")) {
		return fmt.Errorf("not a valid launch file: %s", path)
	}

	cmd, err := os.OpenFile(config.CmdInterface, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer cmd.Close()

	cmd.WriteString(fmt.Sprintf("load_core %s\n", path))

	return nil
}

func launchTempMgl(system *games.System, path string) error {
	mgl, err := GenerateMgl(system, path)
	if err != nil {
		return err
	}

	tmpFile, err := writeTempFile(mgl, "mgl")
	if err != nil {
		return err
	}

	return launchFile(tmpFile)
}

func LaunchGame(system *games.System, path string) error {
	if system == nil {
		return fmt.Errorf("no system specified")
	}

	switch s.ToLower(filepath.Ext(path)) {
	case ".mra":
		err := launchFile(path)
		if err != nil {
			return err
		}
	case ".mgl":
		err := launchFile(path)
		if err != nil {
			return err
		}
	default:
		err := launchTempMgl(system, path)
		if err != nil {
			return err
		}
	}

	if ActiveGameEnabled() {
		SetActiveGame(path)
	}

	return nil
}

func GetLauncherFilename(system *games.System, folder string, name string) string {
	if system.Id == "Arcade" {
		return filepath.Join(folder, name+".mra")
	} else {
		return filepath.Join(folder, name+".mgl")
	}
}

func DeleteLauncher(path string) error {
	if _, err := os.Stat(path); err == nil {
		err := os.Remove(path)
		if err != nil {
			return fmt.Errorf("failed to remove launcher: %s", err)
		}
	}

	// FIXME: best effort for now but this should be case insensitive
	mras, _ := filepath.Glob(filepath.Join(filepath.Dir(path), "*.mra"))
	if len(mras) == 0 {
		coresLink := filepath.Join(filepath.Dir(path), filepath.Base(config.ArcadeCoresFolder))
		if _, err := os.Lstat(coresLink); err == nil {
			err := os.Remove(coresLink)
			if err != nil {
				return fmt.Errorf("failed to remove cores link: %s", err)
			}
		}
	}

	return nil
}

func CreateLauncher(system *games.System, gameFile string, folder string, name string) (string, error) {
	if system == nil {
		return "", fmt.Errorf("no system specified")
	}

	if system.Id == "Arcade" {
		mraPath := GetLauncherFilename(system, folder, name)
		if _, err := os.Lstat(mraPath); err == nil {
			err := os.Remove(mraPath)
			if err != nil {
				return "", fmt.Errorf("failed to remove existing link: %s", err)
			}
		}

		err := os.Symlink(gameFile, mraPath)
		if err != nil {
			return "", fmt.Errorf("failed to create game link: %s", err)
		}

		coresLink := filepath.Join(folder, filepath.Base(config.ArcadeCoresFolder))
		if _, err := os.Lstat(coresLink); err != nil {
			err := os.Symlink(config.ArcadeCoresFolder, coresLink)
			if err != nil {
				return "", fmt.Errorf("failed to create cores link: %s", err)
			}
		}

		return mraPath, nil
	} else {
		mglPath := GetLauncherFilename(system, folder, name)

		mgl, err := GenerateMgl(system, gameFile)
		if err != nil {
			return "", err
		}

		err = os.WriteFile(mglPath, []byte(mgl), 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write mgl file: %s", err)
		}

		return mglPath, nil
	}
}

// Launch a core given a possibly partial path, as per MGL files.
func LaunchCore(path string) error {
	if _, err := os.Stat(config.CmdInterface); err != nil {
		return fmt.Errorf("command interface not accessible: %s", err)
	}

	if !filepath.IsAbs(path) {
		query := filepath.Join(config.SdFolder, path) + "*"
		matches, err := filepath.Glob(query)
		if err != nil {
			return fmt.Errorf("failed to glob for core: %s", err)
		}

		if len(matches) == 0 {
			return fmt.Errorf("no cores found matching: %s", query)
		} else {
			path = matches[0]
		}
	}

	cmd, err := os.OpenFile(config.CmdInterface, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer cmd.Close()

	cmd.WriteString(fmt.Sprintf("load_core %s\n", path))

	return nil
}

func LaunchMenu() error {
	if _, err := os.Stat(config.CmdInterface); err != nil {
		return fmt.Errorf("command interface not accessible: %s", err)
	}

	cmd, err := os.OpenFile(config.CmdInterface, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer cmd.Close()

	cmd.WriteString(fmt.Sprintf("load_core %s\n", filepath.Join(config.SdFolder, "menu.rbf")))

	return nil
}
