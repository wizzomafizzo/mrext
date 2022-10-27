package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// TODO: offer to enable recents option and reboot
// TODO: compatibility with GameEventHub
//       https://github.com/christopher-roelofs/GameEventHub/blob/main/mister.py
// TODO: hashing functions (including inside zips)
// TODO: create example ini file

const pidFile = "/tmp/playlog.pid"
const logFile = "/tmp/playlog.log"

func startService(logger *log.Logger, cfg *config.UserConfig) {
	// TODO: should be a unified lib for managing apps as services
	if _, err := os.Stat(pidFile); err == nil {
		logger.Println("playlog service already running")
		os.Exit(1)
	} else {
		logger.Println("starting playlog service")
		pid := os.Getpid()
		os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
	}

	tr, err := newTracker(logger, cfg)
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
			path, err := filepath.Abs(os.Args[0])
			if err != nil {
				return err
			}

			cmd := fmt.Sprintf("[[ -e %s ]] && %s -service $1", path, path)

			err = startup.Add("mrext/playlog", cmd)
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

	logger := log.New(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    1,
		MaxBackups: 2,
	}, "", log.LstdFlags)

	if !mister.RecentsOptionEnabled() {
		logger.Println("recents option not enabled")
		fmt.Println("The \"recents\" option must be enabled for playlog to work. Configure it in the MiSTer.ini file and reboot.")
		os.Exit(1)
	}

	// TODO: say if an entry was added
	err := tryAddStartup()
	if err != nil {
		logger.Println("error adding startup:", err)
		fmt.Println("Error adding to startup:", err)
	}

	cfg, err := config.LoadUserConfig(config.UserConfig{
		PlayLog: config.PlayLogConfig{
			SaveEvery: 5, // minutes
		},
	})
	if err != nil {
		logger.Println("error loading user config:", err)
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	start := func() {
		err := exec.Command(os.Args[0], "-service", "exec", "&").Start()
		if err != nil {
			logger.Println("error starting service:", err)
			os.Exit(1)
		}
	}

	if *service == "exec" {
		startService(logger, &cfg)
		os.Exit(0)
	} else if *service == "start" {
		start()
		os.Exit(0)
	} else if *service == "stop" {
		stopService(logger)
		os.Exit(0)
	} else if *service == "restart" {
		stopService(logger)
		// TODO: check if this needs a delay
		startService(logger, &cfg)
		os.Exit(0)
	}

	// TODO: more robust way to check if running
	if _, err := os.Stat(pidFile); err != nil {
		start()
	}

	db, err := openPlayLogDb()
	if err != nil {
		logger.Println("error opening db:", err)
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}

	cores, err := db.topCores(10)
	if err != nil {
		logger.Println("error getting top cores:", err)
		fmt.Println("Error getting top cores:", err)
		os.Exit(1)
	}
	maxCoreLen := 0
	for _, core := range cores {
		if len(core.name) > maxCoreLen {
			maxCoreLen = len(core.name)
		}
	}

	games, err := db.topGames(10)
	if err != nil {
		logger.Println("error getting top games:", err)
		fmt.Println("Error getting top games:", err)
		os.Exit(1)
	}
	maxGameLen := 0
	for _, game := range games {
		if len(game.name) > maxGameLen {
			maxGameLen = len(game.name)
		}
	}

	fmt.Println("Top played cores:")
	// TODO: convert names using names.txt
	for _, core := range cores {
		hours := core.time / 3600
		minutes := (core.time % 3600) / 60
		fmt.Printf("%-*s  %dh %dm\n", maxCoreLen, core.name, hours, minutes)
	}
	fmt.Println()
	fmt.Println("Top played games:")
	for _, game := range games {
		hours := game.time / 3600
		minutes := (game.time % 3600) / 60
		fmt.Printf("%-*s  %dh %dm\n", maxGameLen, game.name, hours, minutes)
	}
}
