package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/input"
)

func sendKeyboard(arg string) {
	kb, err := input.NewKeyboard()
	if err != nil {
		fmt.Printf("error creating virtual keyboard: %s", err)
		return
	}
	defer kb.Close()

	var kbdMap = map[string]func(){
		"close":            kb.Close,
		"volumeup":         kb.VolumeUp,
		"volumedown":       kb.VolumeDown,
		"volumemute":       kb.VolumeMute,
		"menu":             kb.Menu,
		"back":             kb.Back,
		"confirm":          kb.Confirm,
		"cancel":           kb.Cancel,
		"up":               kb.Up,
		"down":             kb.Down,
		"left":             kb.Left,
		"right":            kb.Right,
		"osd":              kb.Osd,
		"coreselect":       kb.CoreSelect,
		"screenshot":       kb.Screenshot,
		"rawscreenshot":    kb.RawScreenshot,
		"user":             kb.User,
		"reset":            kb.Reset,
		"pairbluetooth":    kb.PairBluetooth,
		"changebackground": kb.ChangeBackground,
		"togglecoredates":  kb.ToggleCoreDates,
		"console":          kb.Console,
		"computerosd":      kb.ComputerOsd,
	}

	if fn, ok := kbdMap[arg]; ok {
		fn()
	} else if i, err := strconv.Atoi(arg); err == nil {
		err := kb.Device.KeyPress(i)
		if err != nil {
			fmt.Printf("error sending key: %s", err)
		}
	} else {
		fmt.Printf("unknown keyboard command: %s", arg)
		// TODO: print possible commands
	}
}

func main() {
	activePaths := flag.Bool("active-paths", false, "print active system paths")
	allPaths := flag.Bool("all-paths", false, "print all detected system paths")
	sendKb := flag.String("send-keyboard", "", "send keyboard command")
	filterSystems := flag.String("s", "all", "restrict operation to systems (comma separated)")
	timed := flag.Bool("t", false, "show how long operation took")
	flag.Parse()

	start := time.Now()

	var selectedSystems []games.System
	if *filterSystems == "all" {
		selectedSystems = games.AllSystems()
	} else {
		filterIds := strings.Split(*filterSystems, ",")
		for _, filterId := range filterIds {
			system, err := games.LookupSystem(filterId)
			if err != nil {
				continue
			} else {
				selectedSystems = append(selectedSystems, *system)
			}
		}
	}

	if *activePaths {
		paths := games.GetActiveSystemPaths(selectedSystems)
		for _, path := range paths {
			fmt.Printf("%s:%s\n", path.System.Id, path.Path)
		}
	} else if *allPaths {
		paths := games.GetSystemPaths(selectedSystems)
		for _, path := range paths {
			fmt.Printf("%s:%s\n", path.System.Id, path.Path)
		}
	} else if *sendKb != "" {
		sendKeyboard(*sendKb)
	}

	if *timed {
		seconds := int(time.Since(start).Seconds())
		milliseconds := int(time.Since(start).Milliseconds())
		remainder := milliseconds % int(time.Second)
		fmt.Printf("Operation took %d.%ds\n", int(seconds), remainder)
	}

	os.Exit(0)
}
