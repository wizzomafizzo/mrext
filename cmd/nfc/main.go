package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"os"
	"sync"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/utils"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"

	"github.com/clausecker/nfc/v2"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

// TODO: something like the nfc-list utility so new users with unsupported readers can help identify them
// TODO: play a fun sound when a scan is successful or fails
// TODO: a -test command to see what the result of an NDEF would be
// TODO: would it be possible to unlock the OSD with a card?
// TODO: more concrete amiibo support
// TODO: strip colons from UID mapping file entries and make lowercase
// TODO: create a test web nfc reader in separate github repo, hosted on pages
// TODO: way to check the status of the service
// TODO: use a tag to signal that that next tag should have the active game written to it
// TODO: option to use search.db instead of on demand index for random

const (
	appName            = "nfc"
	connectMaxTries    = 10
	timesToPoll        = 20
	periodBetweenPolls = 300 * time.Millisecond
	periodBetweenLoop  = 300 * time.Millisecond
	timeToForgetCard   = 5 * time.Second
)

var (
	supportedCardTypes = []nfc.Modulation{
		{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106},
	}
	logger = service.NewLogger(appName)
)

type Card struct {
	CardType string
	UID      string
	Text     string
	ScanTime time.Time
}

type ServiceState struct {
	mu              sync.Mutex
	activeCard      Card
	lastScanned     Card
	stopService     bool
	disableLauncher bool
	dbLoaded        time.Time
	uidMap          map[string]string
	textMap         map[string]string
}

func (s *ServiceState) SetActiveCard(card Card) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeCard = card
	if s.activeCard.UID != "" {
		s.lastScanned = card
	}
}

func (s *ServiceState) GetActiveCard() Card {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeCard
}

func (s *ServiceState) GetLastScanned() Card {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastScanned
}

func (s *ServiceState) StopService() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopService = true
}

func (s *ServiceState) ShouldStopService() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopService
}

func (s *ServiceState) DisableLauncher() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disableLauncher = true
}

func (s *ServiceState) EnableLauncher() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disableLauncher = false
}

func (s *ServiceState) IsLauncherDisabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disableLauncher
}

func (s *ServiceState) GetDB() (map[string]string, map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.uidMap, s.textMap
}

func (s *ServiceState) GetDBLoaded() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dbLoaded
}

func (s *ServiceState) SetDB(uidMap map[string]string, textMap map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dbLoaded = time.Now()
	s.uidMap = uidMap
	s.textMap = textMap
}

func pollDevice(
	pnd *nfc.Device,
	activeCard Card,
) (Card, error) {
	count, target, err := pnd.InitiatorPollTarget(supportedCardTypes, timesToPoll, periodBetweenPolls)
	if err != nil && !errors.Is(err, nfc.Error(nfc.ETIMEOUT)) {
		return activeCard, fmt.Errorf("error polling: %s", err)
	}

	if count <= 0 {
		if activeCard.UID != "" && time.Since(activeCard.ScanTime) > timeToForgetCard {
			logger.Info("card removed")
			activeCard = Card{}
		}

		return activeCard, nil
	}

	cardUid := getCardUID(target)
	if cardUid == "" {
		logger.Warn("unable to detect card UID: %s", target.String())
	}

	if cardUid == activeCard.UID {
		return activeCard, nil
	}

	logger.Info("card UID: %s", cardUid)

	cardType, err := getCardType(*pnd)
	if err != nil {
		logger.Error("error getting card type: %s", err)
	} else if cardType == "" {
		logger.Warn("unknown card type")
	} else {
		logger.Info("card type: %s", cardType)
	}

	blockCount := getDataAreaSize(cardType)
	record, err := readRecord(*pnd, blockCount)
	if err != nil {
		return activeCard, fmt.Errorf("error reading record: %s", err)
	}
	logger.Info("record bytes: %s", hex.EncodeToString(record))

	tagText := parseRecordText(record)
	if tagText == "" {
		logger.Warn("no text NDEF found")
	} else {
		logger.Info("decoded text NDEF: %s", tagText)
	}

	card := Card{
		CardType: cardType,
		UID:      cardUid,
		Text:     tagText,
		ScanTime: time.Now(),
	}

	return card, nil
}

func startService(cfg *config.UserConfig) (func() error, error) {
	state := &ServiceState{}

	err := loadDatabase(state)
	if err != nil {
		logger.Error("error loading database: %s", err)
	}

	var closeDbWatcher func() error
	dbWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("error creating watcher: %s", err)
	} else {
		closeDbWatcher = dbWatcher.Close
	}

	go func() {
		// this turned out to be not trivial to say the least, mostly due to
		// the fact the fsnotify library does not implement the IN_CLOSE_WRITE
		// inotify event, which signals the file has finished being written
		// see: https://github.com/fsnotify/fsnotify/issues/372
		//
		// during a standard write operation, a file may emit multiple write
		// events, including when the file could be half-written
		//
		//it's also the case that editors may delete the file and create a new
		// one, which kills the active watcher
		//
		// this solution is very ugly, but it appears to work well :)
		// i think it will be sufficient for the use case, and i really like
		// this idea a lot. it's certainly preferable to the screen flicker
		// with the previous setup
		//
		// there doesn't appear to be any actively maintained wrapper for
		// inotify, so i think it would be best to write one for mrext later
		const delay = 1 * time.Second
		for {
			select {
			case event, ok := <-dbWatcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					// usually receives multiple write events, just act on the first
					if time.Since(state.GetDBLoaded()) < delay {
						continue
					}
					time.Sleep(delay)
					logger.Info("database changed, reloading")
					err := loadDatabase(state)
					if err != nil {
						logger.Error("error loading database: %s", err)
					}
				} else if event.Has(fsnotify.Remove) {
					// editors may also delete the file on write
					time.Sleep(delay)
					_, err := os.Stat(config.NfcDatabaseFile)
					if err == nil {
						err = dbWatcher.Add(config.NfcDatabaseFile)
						if err != nil {
							logger.Error("error watching database: %s", err)
						}
						logger.Info("database changed, reloading")
						err := loadDatabase(state)
						if err != nil {
							logger.Error("error loading database: %s", err)
						}
					}
				}
			case err, ok := <-dbWatcher.Errors:
				if !ok {
					return
				}
				logger.Error("watcher error: %s", err)
			}
		}
	}()

	err = dbWatcher.Add(config.NfcDatabaseFile)
	if err != nil {
		logger.Error("error watching database: %s", err)
	}

	go func() {
		var pnd nfc.Device
		var err error

		tries := 0
		for {
			pnd, err = nfc.Open(cfg.NfcConfig.ConnectionString)
			if err != nil {
				logger.Error("could not open device: %s", err)
				if tries >= connectMaxTries {
					logger.Error("giving up, exiting")
					return
				}
			} else {
				break
			}
			tries++
		}

		defer func(pnd nfc.Device) {
			err := pnd.Close()
			if err != nil {
				logger.Warn("error closing device: %s", err)
			}
		}(pnd)

		if err := pnd.InitiatorInit(); err != nil {
			logger.Error("could not init initiator: %s", err)
			return
		}

		logger.Info("opened connection: %s %s", pnd, pnd.Connection())
		logger.Info("polling for %d times with %s delay", timesToPoll, periodBetweenPolls)

		for {
			if state.ShouldStopService() {
				break
			}

			activeCard := state.GetActiveCard()
			newScanned, err := pollDevice(&pnd, activeCard)
			if err != nil {
				logger.Error("error during poll: %s", err)
				goto end
			}

			state.SetActiveCard(newScanned)

			if newScanned.UID == "" || activeCard.UID == newScanned.UID {
				goto end
			}

			err = writeScanResult(newScanned)
			if err != nil {
				logger.Warn("error writing tmp scan result: %s", err)
			}

			err = launchCard(cfg, state)
			if err != nil {
				logger.Error("error launching card: %s", err)
			}

		end:
			time.Sleep(periodBetweenLoop)
		}
	}()

	return func() error {
		state.StopService()
		if closeDbWatcher != nil {
			return closeDbWatcher()
		}
		return nil
	}, nil
}

func writeScanResult(card Card) error {
	f, err := os.Create(config.NfcLastScanFile)
	if err != nil {
		return fmt.Errorf("unable to create scan result file %s: %s", config.NfcLastScanFile, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.WriteString(fmt.Sprintf("%s,%s", card.UID, card.Text))
	if err != nil {
		return fmt.Errorf("unable to write scan result file %s: %s", config.NfcLastScanFile, err)
	}

	return nil
}

func tryAddStartup() error {
	var startup mister.Startup

	err := startup.Load()
	if err != nil {
		return err
	}

	if !startup.Exists("mrext/" + appName) {
		if utils.YesOrNoPrompt("NFC must be set to run on MiSTer startup. Add it now?") {
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
	svcOpt := flag.String("service", "", "manage nfc service (start, stop, restart, status)")
	flag.Parse()

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{})
	if err != nil {
		logger.Error("error loading user config: %s", err)
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	svc, err := service.NewService(service.ServiceArgs{
		Name:   appName,
		Logger: logger,
		Entry: func() (func() error, error) {
			return startService(cfg)
		},
	})
	if err != nil {
		logger.Error("error creating service: %s", err)
		fmt.Println("Error creating service:", err)
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
		fmt.Println("Service is running.")
		os.Exit(0)
	}
}
