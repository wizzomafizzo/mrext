package main

import (
	"flag"
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"os"
)

const appName = "lastplayed"
const defaultName = "Last Played"

func createLastPlayedMgl(cfg *config.UserConfig, path string) error {
	var mglName string

	if cfg.LastPlayed.Name == "" {
		mglName = defaultName
	} else {
		mglName = cfg.LastPlayed.Name
	}

	mglName = utils.StripBadFileChars(mglName)

	systems := games.FolderToSystems(path)
	if len(systems) == 0 {
		return fmt.Errorf("no system match found: %s", path)
	}

	system := systems[0]

	_, err := mister.CreateLauncher(&system, path, config.SdFolder, mglName)
	if err != nil {
		return fmt.Errorf("error creating mgl: %s", err)
	}

	return nil
}

type fakeDb struct {
	config *config.UserConfig
}

func (f *fakeDb) FixPowerLoss() (bool, error) {
	return false, nil
}

func (f *fakeDb) AddEvent(ev tracker.EventAction) error {
	if ev.Action == tracker.EventActionGameStart {
		return createLastPlayedMgl(f.config, ev.TargetPath)
	}

	return nil
}

func (f *fakeDb) UpdateCore(_ tracker.CoreTime) error {
	return nil
}

func (f *fakeDb) GetCore(_ string) (tracker.CoreTime, error) {
	return tracker.CoreTime{}, nil
}

func (f *fakeDb) UpdateGame(_ tracker.GameTime) error {
	return nil
}

func (f *fakeDb) GetGame(_ string) (tracker.GameTime, error) {
	return tracker.GameTime{}, nil
}

func (f *fakeDb) NoResults(_ error) bool {
	return true
}

func startService(logger *service.Logger, cfg *config.UserConfig) (func() error, error) {
	tr, err := tracker.NewTracker(logger, cfg, &fakeDb{
		config: cfg,
	})
	if err != nil {
		logger.Error("error starting tracker: %s", err)
		os.Exit(1)
	}

	tr.LoadCore()
	if !mister.ActiveGameEnabled() {
		err := mister.SetActiveGame("")
		if err != nil {
			tr.Logger.Error("error setting active game: %s", err)
		}
	}

	watcher, err := tracker.StartFileWatch(tr)
	if err != nil {
		tr.Logger.Error("error starting file watch: %s", err)
		os.Exit(1)
	}

	return func() error {
		err := watcher.Close()
		if err != nil {
			tr.Logger.Error("error closing file watcher: %s", err)
		}
		tr.StopAll()
		return nil
	}, nil
}

func tryAddStartup() error {
	var startup mister.Startup

	err := startup.Load()
	if err != nil {
		return err
	}

	if !startup.Exists("mrext/" + appName) {
		if utils.YesOrNoPrompt("LastPlayed must be set to run on MiSTer startup. Add it now?") {
			err = startup.AddService("mrext/" + appName)
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
	svcOpt := flag.String("service", "", "manage playlog service (start, stop, restart, status)")
	flag.Parse()

	logger := service.NewLogger(appName)

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{
		LastPlayed: config.LastPlayedConfig{
			Name: defaultName,
		},
	})
	if err != nil {
		logger.Error("error loading user config: %s", err)
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	svc, err := service.NewService(service.ServiceArgs{
		Name:   appName,
		Logger: logger,
		Entry: func() (func() error, error) {
			return startService(logger, cfg)
		},
	})
	if err != nil {
		logger.Error("error creating service: %s", err)
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}

	if !mister.RecentsOptionEnabled() {
		logger.Error("recents option not enabled, exiting...")
		fmt.Println("The \"recents\" option must be enabled for playlog to work. Configure it in the MiSTer.ini file and reboot.")
		os.Exit(1)
	}

	svc.ServiceHandler(svcOpt)

	err = tryAddStartup()
	if err != nil {
		logger.Error("error adding startup: %s", err)
		fmt.Println("Error adding to startup:", err)
	}

	if !svc.Running() {
		err := svc.Start()
		if err != nil {
			logger.Error("error starting service: %s", err)
			fmt.Println("Error starting service:", err)
			os.Exit(1)
		} else {
			fmt.Println("Service started successfully.")
			os.Exit(0)
		}
	} else {
		fmt.Println("Service is already running.")
		os.Exit(0)
	}
}
