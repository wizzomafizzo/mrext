package games

import (
	"os"
	"path/filepath"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func CopySetnameBios(cfg *config.UserConfig, origSystem System, newSystem System, name string) error {
	var biosPath string

	for _, folder := range GetActiveSystemPaths(cfg, []System{origSystem}) {
		checkPath := filepath.Join(folder.Path, name)
		if _, err := os.Stat(checkPath); err == nil {
			biosPath = checkPath
			break
		}
	}

	if biosPath == "" || newSystem.SetName == "" {
		return nil
	}

	newFolder, err := filepath.Abs(filepath.Join(filepath.Dir(biosPath), "..", newSystem.SetName))
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(newFolder, name)); err == nil {
		return nil
	}

	if err := os.MkdirAll(newFolder, 0755); err != nil {
		return err
	}

	return utils.CopyFile(biosPath, filepath.Join(newFolder, name))
}

func PreHookFDS(cfg *config.UserConfig) error {
	nesSystem, err := GetSystem("NES")
	if err != nil {
		return err
	}

	fdsSystem, err := GetSystem("FDS")
	if err != nil {
		return err
	}

	return CopySetnameBios(cfg, *nesSystem, *fdsSystem, "boot0.rom")
}

func PreHookWSC(cfg *config.UserConfig) error {
	wsSystem, err := GetSystem("WonderSwan")
	if err != nil {
		return err
	}

	wscSystem, err := GetSystem("WonderSwanColor")
	if err != nil {
		return err
	}

	err = CopySetnameBios(cfg, *wsSystem, *wscSystem, "boot.rom")
	if err != nil {
		return err
	}

	return CopySetnameBios(cfg, *wsSystem, *wscSystem, "boot1.rom")
}

var SystemPreHooks = map[string]func(*config.UserConfig) error{
	"FDS":             PreHookFDS,
	"WonderSwanColor": PreHookWSC,
}

func RunSystemPreHook(cfg *config.UserConfig, system System) error {
	if hook, ok := SystemPreHooks[system.Id]; ok {
		return hook(cfg)
	}
	return nil
}
