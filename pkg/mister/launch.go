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
	if path == "" {
		if system.SetName == "" {
			return fmt.Sprintf(
				"<mistergamedescription>\n\t<rbf>%s</rbf>\n</mistergamedescription>\n",
				system.Rbf,
			), nil
		} else {
			return fmt.Sprintf(
				"<mistergamedescription>\n\t<rbf>%s</rbf>\n\t<setname>%s</setname>\n</mistergamedescription>\n",
				system.Rbf, system.SetName,
			), nil
		}
	}

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
	}

	if system.SetName == "" {
		// TODO: generate this from xml
		return fmt.Sprintf(
			"<mistergamedescription>\n\t<rbf>%s</rbf>\n\t<file delay=\"%d\" type=\"%s\" index=\"%d\" path=\"../../../../..%s\"/>\n</mistergamedescription>\n",
			system.Rbf, mglDef.Delay, mglDef.Method, mglDef.Index, path,
		), nil
	} else {
		return fmt.Sprintf(
			"<mistergamedescription>\n\t<rbf>%s</rbf>\n\t<setname>%s</setname>\n\t<file delay=\"%d\" type=\"%s\" index=\"%d\" path=\"../../../../..%s\"/>\n</mistergamedescription>\n",
			system.Rbf, system.SetName, mglDef.Delay, mglDef.Method, mglDef.Index, path,
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

	// TODO: clean up
	if !(s.HasSuffix(s.ToLower(path), ".mgl") || s.HasSuffix(s.ToLower(path), ".mra") || s.HasSuffix(s.ToLower(path), ".rbf")) {
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

func LaunchGame(system games.System, path string) error {
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

		if ActiveGameEnabled() {
			SetActiveGame(path)
		}
	default:
		rbfs := games.SystemsWithRbf()
		if _, ok := rbfs[system.Id]; ok {
			system.Rbf = rbfs[system.Id].MglName
		}

		err := launchTempMgl(&system, path)
		if err != nil {
			return err
		}

		if ActiveGameEnabled() {
			SetActiveGame(path)
		}
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

func TrySetupArcadeCoresLink(path string) error {
	folder, err := os.Stat(path)
	if err != nil {
		return err
	} else if !folder.IsDir() {
		return fmt.Errorf("parent is not a directory: %s", path)
	}

	coresLinkPath := filepath.Join(path, filepath.Base(config.ArcadeCoresFolder))
	coresLink, err := os.Lstat(coresLinkPath)

	coresLinkExists := false
	if err == nil {
		if coresLink.Mode()&os.ModeSymlink != 0 {
			coresLinkExists = true
		} else {
			// cores exists but it's not a symlink. not touching this!
			return nil
		}
	} else if os.IsNotExist(err) {
		coresLinkExists = false
	} else {
		return err
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	mraCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if s.HasSuffix(s.ToLower(file.Name()), ".mra") {
			mraCount++
		}
	}

	if mraCount > 0 && !coresLinkExists {
		err = os.Symlink(config.ArcadeCoresFolder, coresLinkPath)
		if err != nil {
			return err
		}
	} else if mraCount == 0 && coresLinkExists {
		err = os.Remove(coresLinkPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteLauncher(path string) error {
	if _, err := os.Stat(path); err == nil {
		err := os.Remove(path)
		if err != nil {
			return fmt.Errorf("failed to remove launcher: %s", err)
		}
	}

	return TrySetupArcadeCoresLink(filepath.Dir(path))
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

		err = TrySetupArcadeCoresLink(filepath.Dir(mraPath))
		if err != nil {
			return "", err
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

// LaunchCore Launch a core given a possibly partial path, as per MGL files.
func LaunchCore(system games.System) error {
	if _, err := os.Stat(config.CmdInterface); err != nil {
		return fmt.Errorf("command interface not accessible: %s", err)
	}

	if system.SetName != "" {
		return LaunchGame(system, "")
	}

	var path string
	rbfs := games.SystemsWithRbf()
	if _, ok := rbfs[system.Id]; ok {
		path = rbfs[system.Id].Path
	} else {
		return fmt.Errorf("no core found for system %s", system.Id)
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

	// TODO: don't hardcode here
	cmd.WriteString(fmt.Sprintf("load_core %s\n", filepath.Join(config.SdFolder, "menu.rbf")))

	return nil
}

// LaunchGenericFile Given a generic file path, launch it using the correct method, if possible.
func LaunchGenericFile(path string) error {
	file, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path is not accessible: %s", err)
	}

	if file.IsDir() {
		return fmt.Errorf("path is a directory")
	}

	isGame := false
	ext := s.ToLower(filepath.Ext(path))
	switch ext {
	case ".mra":
		err = launchFile(path)
	case ".mgl":
		err = launchFile(path)
		isGame = true
	case ".rbf":
		err = launchFile(path)
	default:
		system, err := games.BestSystemMatch(path)
		if err != nil {
			return fmt.Errorf("unknown file type: %s", ext)
		}

		err = launchTempMgl(&system, path)
		isGame = true
	}

	if err != nil {
		return err
	}

	if ActiveGameEnabled() && isGame {
		err := SetActiveGame(path)
		if err != nil {
			return err
		}
	}

	return nil
}
