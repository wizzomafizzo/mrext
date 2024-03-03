package games

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func copySetnameBios(cfg *config.UserConfig, origSystem System, newSystem System, name string) error {
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

func hookFDS(cfg *config.UserConfig, system System, _ string) (string, error) {
	nesSystem, err := GetSystem("NES")
	if err != nil {
		return "", err
	}

	return "", copySetnameBios(cfg, *nesSystem, system, "boot0.rom")
}

func hookWSC(cfg *config.UserConfig, system System, _ string) (string, error) {
	wsSystem, err := GetSystem("WonderSwan")
	if err != nil {
		return "", err
	}

	err = copySetnameBios(cfg, *wsSystem, system, "boot.rom")
	if err != nil {
		return "", err
	}

	return "", copySetnameBios(cfg, *wsSystem, system, "boot1.rom")
}

func hookAo486(_ *config.UserConfig, system System, path string) (string, error) {
	mglDef, err := PathToMglDef(system, path)
	if err != nil {
		return "", err
	}

	var mgl string

	if !strings.HasSuffix(strings.ToLower(path), ".vhd") {
		return "", nil
	}

	dir := filepath.Dir(path)
	filename := filepath.Base(path)

	// exception for Top 300 pack which uses 2 disks
	if strings.HasSuffix(path, "IDE 0-1 Top 300 DOS Games.vhd") {
		mgl += fmt.Sprintf(
			"\t<file delay=\"%d\" type=\"%s\" index=\"%d\" path=\"%s\"/>\n",
			mglDef.Delay,
			mglDef.Method,
			mglDef.Index,
			filepath.Join(dir, "IDE 0-0 BOOT-DOS98.vhd"),
		)

		mgl += fmt.Sprintf(
			"\t<file delay=\"%d\" type=\"%s\" index=\"%d\" path=\"%s\"/>\n",
			mglDef.Delay,
			mglDef.Method,
			mglDef.Index+1,
			path,
		)

		mgl += "\t<reset delay=\"1\"/>\n"

		return mgl, nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	// if there's an iso in the same folder, mount it too
	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file.Name()), ".iso") && file.Name() != filename {
			mgl += fmt.Sprintf(
				"\t<file delay=\"%d\" type=\"%s\" index=\"%d\" path=\"%s\"/>\n",
				mglDef.Delay,
				mglDef.Method,
				4,
				filepath.Join(dir, file.Name()),
			)
			break
		}
	}

	mgl += fmt.Sprintf(
		"\t<file delay=\"%d\" type=\"%s\" index=\"%d\" path=\"%s\"/>\n",
		mglDef.Delay,
		mglDef.Method,
		mglDef.Index,
		path,
	)

	mgl += "\t<reset delay=\"1\"/>\n"

	return mgl, nil
}

func hookAmiga(_ *config.UserConfig, system System, path string) (string, error) {
	if !strings.HasSuffix(strings.ToLower(filepath.Dir(path)), "listings/games.txt") && !strings.HasSuffix(strings.ToLower(filepath.Dir(path)), "listings/demos.txt") {
		return "", nil
	}

	gameName := filepath.Base(path)
	sharedPath, err := filepath.Abs(filepath.Join(filepath.Dir(path), "..", "..", "shared"))
	if err != nil {
		return "", err
	}

	bootFile := filepath.Join(sharedPath, "ags_boot")
	if err := os.WriteFile(bootFile, []byte(gameName+"\n"), 0644); err != nil {
		return "", err
	}

	return "\t<setname>Amiga</setname>\n", nil
}

func hookNeoGeo(_ *config.UserConfig, _ System, path string) (string, error) {
	// neogeo core allows launching zips and folders
	if strings.HasSuffix(strings.ToLower(path), ".zip") || filepath.Ext(path) == "" {
		return fmt.Sprintf(
			"\t<file delay=\"%d\" type=\"%s\" index=\"%d\" path=\"%s\"/>\n",
			1,
			"f",
			1,
			path,
		), nil
	}

	return "", nil
}

var systemHooks = map[string]func(*config.UserConfig, System, string) (string, error){
	"FDS":             hookFDS,
	"WonderSwanColor": hookWSC,
	"ao486":           hookAo486,
	"Amiga":           hookAmiga,
	"NeoGeo":          hookNeoGeo,
}

func RunSystemHook(cfg *config.UserConfig, system System, path string) (string, error) {
	if hook, ok := systemHooks[system.Id]; ok {
		return hook(cfg, system, path)
	}

	return "", nil
}
