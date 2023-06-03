package main

import (
	"flag"
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/mister"
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
	getIni := flag.Bool("get-ini", false, "get active ini file")
	setIni := flag.Int("set-ini", -1, "set active ini file (1-4)")
	listInis := flag.Bool("list-inis", false, "list available ini files")
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
	} else if *getIni {
		n, err := mister.GetCurrentIni()
		if err != nil {
			fmt.Printf("error getting ini: %s\n", err)
			os.Exit(1)
		}
		if n == 0 {
			fmt.Printf("no active ini\n")
			os.Exit(0)
		}
		fmt.Printf("active ini: %d\n", n)
	} else if *setIni != -1 {
		err := mister.SetCurrentIni(*setIni)
		if err != nil {
			fmt.Printf("error setting ini: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("set active ini to %d\n", *setIni)
	} else if *listInis {
		inis, err := mister.ListMisterInis()
		if err != nil {
			fmt.Printf("error listing inis: %s\n", err)
			os.Exit(1)
		}

		for i, ini := range inis {
			fmt.Printf("%d: %s (%s)\n", i+1, ini.DisplayName, ini.Filename)
		}
	} else {
		flag.Usage()
	}

	if *timed {
		seconds := int(time.Since(start).Seconds())
		milliseconds := int(time.Since(start).Milliseconds())
		remainder := milliseconds % int(time.Second)
		fmt.Printf("Operation took %d.%ds\n", int(seconds), remainder)
	}

	os.Exit(0)
}
