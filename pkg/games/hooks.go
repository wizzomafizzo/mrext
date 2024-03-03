package games

import (
	"os"
	"path/filepath"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// Create the FDS folder and copy over the FDS BIOS from the NES folder if possible.
func PreHookFDS(cfg *config.UserConfig) error {
	const bootName = "boot0.rom"
	var biosPath string

	nesSystem, err := GetSystem("NES")
	if err != nil {
		return err
	}

	for _, folder := range GetActiveSystemPaths(cfg, []System{*nesSystem}) {
		checkPath := filepath.Join(folder.Path, bootName)
		if _, err := os.Stat(checkPath); err == nil {
			biosPath = checkPath
			break
		}
	}

	if biosPath == "" {
		return nil
	}

	fdsFolder, err := filepath.Abs(filepath.Join(filepath.Dir(biosPath), "..", "FDS"))
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(fdsFolder, bootName)); err == nil {
		return nil
	}

	if err := os.MkdirAll(fdsFolder, 0755); err != nil {
		return err
	}

	return utils.CopyFile(biosPath, filepath.Join(fdsFolder, bootName))
}

var SystemPreHooks = map[string]func(*config.UserConfig) error{
	"FDS": PreHookFDS,
}

func RunSystemPreHook(cfg *config.UserConfig, system System) error {
	if hook, ok := SystemPreHooks[system.Id]; ok {
		return hook(cfg)
	}
	return nil
}
