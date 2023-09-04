package mister

import (
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	s "strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
)

func GenerateMgl(cfg *config.UserConfig, system *games.System, path string) (string, error) {
	// override the system rbf with the user specified one
	for _, setCore := range cfg.Systems.SetCore {
		parts := s.SplitN(setCore, ":", 2)
		if len(parts) != 2 {
			continue
		}

		if s.EqualFold(parts[0], system.Id) {
			system.Rbf = parts[1]
			break
		}
	}

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

func launchTempMgl(cfg *config.UserConfig, system *games.System, path string) error {
	mgl, err := GenerateMgl(cfg, system, path)
	if err != nil {
		return err
	}

	tmpFile, err := writeTempFile(mgl, "mgl")
	if err != nil {
		return err
	} else {
		go func() {
			time.Sleep(5 * time.Second)
			_ = os.Remove(tmpFile)
		}()
	}

	return launchFile(tmpFile)
}

// LaunchShortCore attempts to launch a core with a short path, as per what's
// allowed in an MGL file.
func LaunchShortCore(path string) error {
	mgl := fmt.Sprintf(
		"<mistergamedescription>\n\t<rbf>%s</rbf>\n</mistergamedescription>\n",
		path,
	)

	tmpFile, err := writeTempFile(mgl, "mgl")
	if err != nil {
		return err
	}

	return launchFile(tmpFile)
}

func LaunchGame(cfg *config.UserConfig, system games.System, path string) error {
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

		err := launchTempMgl(cfg, &system, path)
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

func CreateLauncher(cfg *config.UserConfig, system *games.System, gameFile string, folder string, name string) (string, error) {
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

		mgl, err := GenerateMgl(cfg, system, gameFile)
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
func LaunchCore(cfg *config.UserConfig, system games.System) error {
	if _, err := os.Stat(config.CmdInterface); err != nil {
		return fmt.Errorf("command interface not accessible: %s", err)
	}

	if system.SetName != "" {
		return LaunchGame(cfg, system, "")
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
func LaunchGenericFile(cfg *config.UserConfig, path string) error {
	// TODO: move this to a common function, and the one in launchtoken
	inZip := false
	parts := s.Split(path, "/")
	for i, part := range parts {
		if s.HasSuffix(s.ToLower(part), ".zip") {
			zipPath := filepath.Join(parts[:i+1]...)
			if path[0] == '/' {
				zipPath = "/" + zipPath
			}
			if _, err := os.Stat(zipPath); err != nil {
				return fmt.Errorf("containing zip file not found for %s: %s", path, zipPath)
			}
			inZip = true
			break
		}
	}

	if !inZip {
		file, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("path is not accessible: %s", err)
		}

		if file.IsDir() {
			return fmt.Errorf("path is a directory")
		}
	}

	var err error
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
		system, err := games.BestSystemMatch(cfg, path)
		if err != nil {
			return fmt.Errorf("unknown file type: %s", ext)
		}

		err = launchTempMgl(cfg, &system, path)
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

// TryPickRandomGame recursively searches through given folder for a valid game
// file for that system.
func TryPickRandomGame(system *games.System, folder string) (string, error) {
	files, err := os.ReadDir(folder)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no files in %s", folder)
	}

	var validFiles []os.DirEntry
	for _, file := range files {
		if file.IsDir() {
			validFiles = append(validFiles, file)
		} else if utils.IsZip(file.Name()) {
			validFiles = append(validFiles, file)
		} else if games.MatchSystemFile(*system, file.Name()) {
			validFiles = append(validFiles, file)
		}
	}

	if len(validFiles) == 0 {
		return "", fmt.Errorf("no valid files in %s", folder)
	}

	file, err := utils.RandomElem(validFiles)
	if err != nil {
		return "", err
	}

	path := filepath.Join(folder, file.Name())
	if file.IsDir() {
		return TryPickRandomGame(system, path)
	} else if utils.IsZip(path) {
		// zip files
		zipFiles, err := utils.ListZip(path)
		if err != nil {
			return "", err
		}
		if len(zipFiles) == 0 {
			return "", fmt.Errorf("no files in %s", path)
		}
		// just shoot our shot on a zip instead of checking every file
		randomZip, err := utils.RandomElem(zipFiles)
		if err != nil {
			return "", err
		}
		zipPath := filepath.Join(path, randomZip)
		if games.MatchSystemFile(*system, zipPath) {
			return zipPath, nil
		} else {
			return "", fmt.Errorf("invalid file picked in %s", path)
		}
	} else {
		return path, nil
	}
}

func LaunchRandomGame(cfg *config.UserConfig, systems []games.System) error {
	const maxTries = 100

	populated := games.GetPopulatedGamesFolders(cfg, systems)
	if len(populated) == 0 {
		return fmt.Errorf("no populated games folders found")
	}

	for i := 0; i < maxTries; i++ {
		systemId, err := utils.RandomElem(utils.MapKeys(populated))
		if err != nil {
			return err
		}

		folders := populated[systemId]
		var files []string
		for _, folder := range folders {
			results, err := games.GetFiles(systemId, folder)
			if err != nil {
				return err
			}
			files = append(files, results...)
		}

		if len(files) == 0 {
			continue
		}

		system, err := games.GetSystem(systemId)
		if err != nil {
			return err
		}

		game, err := utils.RandomElem(files)
		if err != nil {
			return err
		}

		return LaunchGame(cfg, *system, game)
	}

	return fmt.Errorf("failed to find a random game")
}

func LaunchToken(cfg *config.UserConfig, manual bool, text string) error {
	// detection can never be perfect, but these characters are illegal in
	// windows filenames and heavily avoided in linux. use them to mark that
	// this is a command
	if s.HasPrefix(text, "**") {
		text = s.TrimPrefix(text, "**")
		parts := s.SplitN(text, ":", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid command: %s", text)
		}

		cmd, args := s.TrimSpace(parts[0]), s.TrimSpace(parts[1])

		// TODO: search game file
		// TODO: game file by hash

		switch cmd {
		case "system":
			if s.EqualFold(args, "menu") {
				return LaunchMenu()
			}

			system, err := games.LookupSystem(args)
			if err != nil {
				return err
			}

			return LaunchCore(cfg, *system)
		case "command":
			if !manual {
				return fmt.Errorf("commands must be manually run")
			}

			command := exec.Command("bash", "-c", args)
			err := command.Start()
			if err != nil {
				return err
			}

			return nil
		case "random":
			if args == "" {
				return fmt.Errorf("no system specified")
			}

			if args == "all" {
				return LaunchRandomGame(cfg, games.AllSystems())
			}

			// TODO: allow multiple systems
			system, err := games.LookupSystem(args)
			if err != nil {
				return err
			}

			return LaunchRandomGame(cfg, []games.System{*system})
		default:
			return fmt.Errorf("unknown command: %s", cmd)
		}
	}

	// if it's not a command, assume it's some kind of file path
	if filepath.IsAbs(text) {
		return LaunchGenericFile(cfg, text)
	}

	// if it's a relative path with no extension, assume it's a core
	if filepath.Ext(text) == "" {
		return LaunchShortCore(text)
	}

	// if the file is in a .zip, just check .zip exists in each games folder
	parts := s.Split(text, "/")
	for i, part := range parts {
		if s.HasSuffix(s.ToLower(part), ".zip") {
			zipPath := filepath.Join(parts[:i+1]...)
			for _, folder := range games.GetGamesFolders(cfg) {
				if _, err := os.Stat(filepath.Join(folder, zipPath)); err == nil {
					return LaunchGenericFile(cfg, filepath.Join(folder, text))
				}
			}
			break
		}
	}

	// then try check for the whole path in each game folder
	for _, folder := range games.GetGamesFolders(cfg) {
		path := filepath.Join(folder, text)
		if _, err := os.Stat(path); err == nil {
			return LaunchGenericFile(cfg, path)
		}
	}

	return fmt.Errorf("could not find file: %s", text)
}
