package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// TODO: offer to enable recents option and reboot
// TODO: compatibility with GameEventHub
//       https://github.com/christopher-roelofs/GameEventHub/blob/main/mister.py
// TODO: hashing functions (including inside zips)
// TODO: disbale write interval with 0

const pidFile = "/tmp/playlog.pid"
const logFile = "/tmp/playlog.log"

func startService(logger *log.Logger, cfg config.UserConfig) {
	// TODO: should be a unified lib for managing apps as services
	if _, err := os.Stat(pidFile); err == nil {
		logger.Println("playlog service already running")
		os.Exit(1)
	} else {
		logger.Println("starting playlog service")
		pid := os.Getpid()
		os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
	}

	tr, err := newTracker(logger)
	if err != nil {
		logger.Println("error starting tracker:", err)
		os.Exit(1)
	}

	// TODO: and this, move to separate lib
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		tr.logger.Println("stopping playlog service")
		tr.stopAll()
		os.Remove(pidFile)
		os.Exit(0)
	}()

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
		interval = 0
	}
	tr.startTicker(interval)

	<-make(chan struct{})
}

func stopService(logger *log.Logger) {
	if _, err := os.Stat(pidFile); err == nil {
		pid, err := os.ReadFile(pidFile)
		if err != nil {
			logger.Println("error reading pid file:", err)
			os.Exit(1)
		}

		pidInt, err := strconv.Atoi(string(pid))
		if err != nil {
			logger.Println("error parsing pid:", err)
			os.Exit(1)
		}

		err = syscall.Kill(pidInt, syscall.SIGTERM)
		if err != nil {
			logger.Println("error stopping service:", err)
			os.Exit(1)
		}
	} else {
		logger.Println("playlog service not running")
	}
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
	service := flag.String("service", "", "manage playlog service (start, stop, restart)")
	flag.Parse()

	// TODO: log to file if debug option active
	lf, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		os.Exit(1)
	}
	defer lf.Close()
	writer := io.MultiWriter(os.Stdout, lf)
	logger := log.New(writer, "", log.LstdFlags)

	// TODO: these errors should be logged to file
	if !mister.RecentsOptionEnabled() {
		fmt.Println("The \"recents\" option must be enabled for playlog to work. Configure it in the MiSTer.ini file and reboot.")
		os.Exit(1)
	}

	err = tryAddStartup()
	if err != nil {
		fmt.Println("Error adding to startup:", err)
	}

	// TODO: log user config to file
	cfg, err := config.LoadUserConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	if *service == "exec" {
		startService(logger, cfg)
		os.Exit(0)
	} else if *service == "start" {
		err := exec.Command(os.Args[0], "-service", "exec", "&").Start()
		if err != nil {
			fmt.Println("Error starting service:", err)
			os.Exit(1)
		}
		os.Exit(0)
	} else if *service == "stop" {
		stopService(logger)
		os.Exit(0)
	} else if *service == "restart" {
		stopService(logger)
		// TODO: check if this needs a delay
		startService(logger, cfg)
		os.Exit(0)
	}
}
