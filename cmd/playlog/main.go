package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// TODO: offer to enable recents option and reboot
// TODO: handle failed mgl launch
// TODO: ticker interval and save interval should be configurable
// TODO: fix event log after power loss
// TODO: enable logging to file
// TODO: compatibility with GameEventHub
//       https://github.com/christopher-roelofs/GameEventHub/blob/main/mister.py
// TODO: hashing functions (including inside zips)

func startService() {
	tr, err := newTracker()
	if err != nil {
		tr.logger.Println("error opening database:", err)
		os.Exit(1)
	}

	tr.loadCore()
	if !mister.ActiveGameEnabled() {
		mister.SetActiveGame("")
	}

	watcher, err := startFileWatch(tr)
	if err != nil {
		tr.logger.Println("error starting file watch:", err)
		os.Exit(1)
	}
	defer watcher.Close()

	tr.startTicker()

	<-make(chan struct{})
}

func tryAddStartup() error {
	var startup mister.Startup

	err := startup.Load()
	if err != nil {
		return err
	}

	if !startup.Exists("mrext/playlog") {
		if utils.YesOrNoPrompt("Play Log must be set to run on MiSTer startup. Add it now?") {
			// TODO: prefer not to hardcode the path
			path := "/media/fat/Scripts/playlog.sh"
			cmd := fmt.Sprintf("[[ -e %s ]] && %s -service $1", path, path)

			err := startup.Add("mrext/playlog", cmd)
			if err != nil {
				return err
			}

			err = startup.Save()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	service := flag.String("service", "", "manage playlog service")
	flag.Parse()

	if !mister.RecentsOptionEnabled() {
		fmt.Println("The \"recents\" option must be enabled for Play Log to work. Configure it in the MiSTer.ini file and reboot.")
		os.Exit(1)
	}

	err := tryAddStartup()
	if err != nil {
		fmt.Println("Error adding to startup:", err)
	}

	if *service == "start" {
		startService()
		os.Exit(0)
	}

}
