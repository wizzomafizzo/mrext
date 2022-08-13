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

func generateMgl(system *games.System, path string) (string, error) {
	var mglDef *games.MglParams

	for _, ft := range system.FileTypes {
		for _, ext := range ft.Extensions {
			if s.HasSuffix(s.ToLower(path), ext) {
				mglDef = ft.Mgl
			}
		}
	}

	if mglDef == nil {
		return "", fmt.Errorf("system has no matching mgl args: %s, %s", system.Id, path)
	} else {
		return fmt.Sprintf(
			"<mistergamedescription>\n\t<rbf>%s</rbf>\n\t<file delay=\"%d\" type=\"%s\" index=\"%d\" path=\"../../../..%s\"/>\n</mistergamedescription>\n",
			system.Rbf, mglDef.Delay, mglDef.FileType, mglDef.Index, path,
		), nil
	}
}

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
	_, err := os.Stat(config.CMD_INTERFACE)
	if err != nil {
		return fmt.Errorf("command interface not accessible: %s", err)
	}

	if !(s.HasSuffix(s.ToLower(path), ".mgl") || s.HasSuffix(s.ToLower(path), ".mra")) {
		return fmt.Errorf("not a valid launch file: %s", path)
	}

	cmd, err := os.OpenFile(config.CMD_INTERFACE, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer cmd.Close()

	cmd.WriteString(fmt.Sprintf("load_core %s\n", path))

	return nil
}

func launchTempMgl(system *games.System, path string) error {
	mgl, err := generateMgl(system, path)
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
		return fmt.Errorf("unknown system: %s", path)
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

	return nil
}
