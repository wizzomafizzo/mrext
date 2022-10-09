package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// TODO: offer to enable recents option and reboot
// TODO: handle failed mgl launch
// TODO: fix event log after power loss
// TODO: enable logging to file
// TODO: compatibility with GameEventHub
//       https://github.com/christopher-roelofs/GameEventHub/blob/main/mister.py
// TODO: hashing functions (including inside zips)

func startService(logger *log.Logger, cfg config.UserConfig) {
	tr, err := newTracker(logger)
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

	var interval int
	if cfg.PlayLog.SaveEvery > 0 {
		interval = cfg.PlayLog.SaveEvery
	} else {
		interval = defaultSaveInterval
	}
	tr.startTicker(interval)

	<-make(chan struct{})
}

func tryAddStartup() error {
	var startup mister.Startup

	err := startup.Load()
	if err != nil {
		return err
	}

	if !startup.Exists("mrext/playlog") {
		if utils.YesOrNoPrompt("PlayLog must be set to run on MiSTer startup. Add it now?") {
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

	// TODO: log to file on -debug
	logger := log.New(os.Stdout, "", log.LstdFlags)

	if !mister.RecentsOptionEnabled() {
		fmt.Println("The \"recents\" option must be enabled for playlog to work. Configure it in the MiSTer.ini file and reboot.")
		os.Exit(1)
	}

	err := tryAddStartup()
	if err != nil {
		fmt.Println("Error adding to startup:", err)
	}

	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	if *service == "start" {
		startService(logger, cfg)
		os.Exit(0)
	}

}
