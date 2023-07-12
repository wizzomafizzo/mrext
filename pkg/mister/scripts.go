package mister

import (
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Script struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
}

func IsMenuRunning() bool {
	activeCore, err := GetActiveCoreName()
	if err != nil {
		return false
	}

	return activeCore == config.MenuCore
}

func IsScriptRunning() bool {
	cmd := "ps ax | grep /tmp/script | grep -v grep"
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return false
	}

	if len(out) > 0 {
		return true
	} else {
		return false
	}
}

func KillActiveScript() error {
	if !IsScriptRunning() {
		return nil
	}

	// TODO: this doesn't actually work right now. it just orphans the launched script process
	// one good idea is to launch scripts with and env variable that contains the pid of the menu
	// so it will get picked up in the grep. it's not urgent though

	cmd := "ps ax | grep /tmp/script | grep -v grep | awk '{print $1}' | xargs kill"
	return exec.Command("sh", "-c", cmd).Run()
}

func ScriptCanLaunch() bool {
	scriptRunning := IsScriptRunning()

	if IsMenuRunning() && !scriptRunning {
		return true
	} else {
		return false
	}
}

func OpenConsole(kbd input.Keyboard) error {
	if !IsMenuRunning() {
		return fmt.Errorf("cannot open console, active core is not menu")
	}

	getTty := func() (string, error) {
		sys := "/sys/devices/virtual/tty/tty0/active"
		if _, err := os.Stat(sys); err != nil {
			return "", err
		}

		tty, err := os.ReadFile(sys)
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(tty)), nil
	}

	// we use the F9 key as a means to disable main's usage of the framebuffer and allow scripts to run
	// unfortunately when the menu "sleeps", any key press will be eaten by main and not trigger the console switch
	// there's also no simple way to tell if mister has switched to the console
	// so what we do is switch to tty3, which is unused by mister, then attempt to switch to console,
	// which sets tty to 1 on success, then check in a loop if it actually did change to 1 and keep pressing F9
	// until it's switched

	err := exec.Command("chvt", "3").Run()
	if err != nil {
		return err
	}

	tries := 0
	tty := ""
	for {
		if tries > 20 {
			return fmt.Errorf("could not switch to tty1")
		}
		kbd.Console()
		time.Sleep(50 * time.Millisecond)
		tty, err = getTty()
		if err != nil {
			return err
		}
		if tty == "tty1" {
			break
		}
		tries++
	}

	return nil
}

func GetAllScripts() ([]Script, error) {
	scripts := make([]Script, 0)

	files, err := os.ReadDir(config.ScriptsFolder)
	if err != nil {
		return scripts, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fn := file.Name()
		if strings.HasSuffix(strings.ToLower(fn), ".sh") {
			scripts = append(scripts, Script{
				Name:     strings.TrimSuffix(fn, filepath.Ext(fn)),
				Filename: fn,
				Path:     filepath.Join(config.ScriptsFolder, fn),
			})
		}
	}

	return scripts, nil
}

func RunScript(kbd input.Keyboard, path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}

	canLaunch := ScriptCanLaunch()
	if !canLaunch {
		return fmt.Errorf("script cannot be launched, active core is not menu or script is already running")
	}

	err := OpenConsole(kbd)
	if err != nil {
		return err
	}

	// this is just to follow mister's convention, which reserves tty2 for scripts
	err = exec.Command("chvt", "2").Run()
	if err != nil {
		return err
	}

	// this is how mister launches scripts itself
	// TODO: press any key should be configurable
	launcher := fmt.Sprintf(`#!/bin/bash
export LC_ALL=en_US.UTF-8
export HOME=/root
export LESSKEY=/media/fat/linux/lesskey
cd $(dirname "%s")
%s
`, path, path)

	// TODO: this is no longer functional, if we still even want it, need to find a way to make it wait for
	//       input but not block in the background like for the random script
	//echo "Press any key to continue"
	//read -n 1 -s -r -p ""

	err = os.WriteFile("/tmp/script", []byte(launcher), 0755)
	if err != nil {
		return err
	}

	err = exec.Command(
		"/sbin/agetty",
		"-a",
		"root",
		"-l",
		"/tmp/script",
		"--nohostname",
		"-L",
		"tty2",
		"linux",
	).Run()
	if err != nil {
		return err
	}

	kbd.ExitConsole()

	return nil
}
