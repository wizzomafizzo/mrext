package main

import (
	_ "embed"
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"os"
	"os/exec"
	"path/filepath"
)

const playerDir = "/tmp/mrext-mplayer"

//go:embed _player/mplayer
var mplayer []byte

//go:embed _player/libtinfo.so.6
var libtinfo []byte

func setupPlayer() error {
	if _, err := os.Stat(playerDir); os.IsNotExist(err) {
		err := os.Mkdir(playerDir, 0755)
		if err != nil {
			return err
		}
	}

	playerPath := filepath.Join(playerDir, "mplayer")
	if _, err := os.Stat(playerPath); os.IsNotExist(err) {
		err := os.WriteFile(playerPath, mplayer, 0755)
		if err != nil {
			return err
		}
	}

	libtinfoPath := filepath.Join(playerDir, "libtinfo.so.6")
	if _, err := os.Stat(libtinfoPath); os.IsNotExist(err) {
		err := os.WriteFile(libtinfoPath, libtinfo, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func setVirtualTerm(id string) error {
	cmd := exec.Command("chvt", id)
	return cmd.Run()
}

func writeTty(id string, s string) error {
	tty := "/dev/tty" + id

	f, err := os.OpenFile(tty, os.O_WRONLY, 0)
	if err != nil {
		return err
	}

	_, err = f.WriteString(s)
	if err != nil {
		return err
	}

	return f.Close()
}

func hideCursor(vt string) error {
	return writeTty(vt, "\033[?25l")
}

func showCursor(vt string) error {
	return writeTty(vt, "\033[?25h")
}

func setupRemotePlay() error {
	kbd, err := input.NewKeyboard()
	if err != nil {
		return fmt.Errorf("error opening keyboard: %w", err)
	}
	defer kbd.Close()

	err = setVirtualTerm("9")
	if err != nil {
		return fmt.Errorf("error switching to virtual terminal 9: %w", err)
	}

	err = hideCursor("9")
	if err != nil {
		return fmt.Errorf("error hiding cursor: %w", err)
	}

	kbd.Console()

	err = mister.SetVideoMode(640, 480)
	if err != nil {
		return fmt.Errorf("error setting video mode: %w", err)
	}

	return nil
}

func cleanupRemotePlay() error {
	err := showCursor("9")
	if err != nil {
		return fmt.Errorf("error showing cursor: %w", err)
	}

	err = setVirtualTerm("1")
	if err != nil {
		return fmt.Errorf("error switching to virtual terminal 1: %w", err)
	}

	err = mister.LaunchMenu()
	if err != nil {
		return fmt.Errorf("error launching menu: %w", err)
	}

	return nil
}

func runMplayer(path string) error {
	cmd := exec.Command(filepath.Join(playerDir, "mplayer"), path)
	cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+playerDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	err := setupPlayer()
	if err != nil {
		panic(err)
	}
}

func main() {
	//err := setupRemotePlay()
	//if err != nil {
	//	panic(err)
	//}

	err := runMplayer(os.Args[1])
	if err != nil {
		panic(err)
	}

	// TODO: handle signals like ctrl-c here
	//err = cleanupRemotePlay()
	//if err != nil {
	//	panic(err)
	//}
}
