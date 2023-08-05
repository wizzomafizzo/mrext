package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/utils"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"

	"github.com/clausecker/nfc/v2"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

// TODO: something like the nfc-list utility so new users with unsupported readers can help identify them

var (
	appName            = "nfc"
	supportedCardTypes = []nfc.Modulation{
		{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106},
	}
	connectMaxTries    = 10
	timesToPoll        = 20
	periodBetweenPolls = 300 * time.Millisecond
	periodBetweenLoop  = 300 * time.Millisecond
	databaseFile       = filepath.Join(config.SdFolder, "nfc-mapping.csv")
	lastScanFile       = filepath.Join(config.TempFolder, "NFCSCAN")
)

type lastSeenCard struct {
	CardType string
	UID      string
	ScanTime time.Time
}

func pollDevice(
	logger *service.Logger,
	cfg *config.UserConfig,
	pnd *nfc.Device,
	lastSeen lastSeenCard,
	db *map[string]string,
) (lastSeenCard, error) {
	count, target, err := pnd.InitiatorPollTarget(supportedCardTypes, timesToPoll, periodBetweenPolls)
	if err != nil && !errors.Is(err, nfc.Error(nfc.ETIMEOUT)) {
		return lastSeen, fmt.Errorf("error polling: %s", err)
	}

	if count <= 0 {
		return lastSeen, nil
	}

	currentCardID := getCardUID(target)
	if currentCardID == "" {
		logger.Warn("unsupported card type: %s", target.String())
	}

	// TODO: i'd like to put this check on on a timer so you can still have cards that are
	//       meant to be scanned multiple times in a row
	if currentCardID == lastSeen.UID {
		return lastSeen, nil
	}

	logger.Info("new card UID: %s", currentCardID)

	cardType, err := getCardType(*pnd)
	if err != nil {
		logger.Error("error getting card type: %s", err)
	}

	if cardType == "" {
		logger.Warn("unknown card type")
	} else {
		logger.Info("card type: %s", cardType)
	}

	blockCount := getDataAreaSize(cardType)
	record, err := readRecord(*pnd, blockCount)
	if err != nil {
		return lastSeenCard{}, fmt.Errorf("error reading record: %s", err)
	}
	logger.Info("record bytes: %s", hex.EncodeToString(record))

	tagText := parseRecordText(record)

	if tagText != "" {
		logger.Info("decoded text NDEF: %s", tagText)

		err = writeScanResult(tagText)
		if err != nil {
			logger.Warn("error writing tmp scan result: %s", err)
		}

		err = loadCoreFromFilename(cfg, tagText)
		if err != nil {
			logger.Error("error loading core: %s", err)
		}
		// TODO: if string is in special format
		//       e.g. !!GBA:abad9c764c35b8202e3d9e5915ca7007bdc7cc62 try to load that way.
	} else {
		logger.Info("no text NDEF found, falling back to UID mapping in CSV file")
		err = loadCoreFromCardUID(cfg, *db, currentCardID)
		if err != nil {
			logger.Error("error loading core: %s", err)
		}
	}

	return lastSeenCard{
		CardType: cardType,
		UID:      currentCardID,
		ScanTime: time.Now(),
	}, nil
}

func startService(logger *service.Logger, cfg *config.UserConfig) (func() error, error) {
	var stopService bool
	go func() {
		logger.Info("loading database: %s", databaseFile)
		database, err := loadDatabase()
		if err != nil {
			logger.Error("error loading database: %s", err)
		} else {
			logger.Info("loaded %d mappings", len(database))
		}

		var pnd nfc.Device
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

		lastSeen := lastSeenCard{}

		for {
			if stopService {
				break
			}

			newSeen, err := pollDevice(logger, cfg, &pnd, lastSeen, &database)
			if err != nil {
				logger.Error("error during poll: %s", err)
			} else {
				lastSeen = newSeen
			}

			time.Sleep(periodBetweenLoop)
		}
	}()

	return func() error {
		stopService = true
		return nil
	}, nil
}

func writeScanResult(tagText string) error {
	f, err := os.Create(lastScanFile)
	if err != nil {
		return fmt.Errorf("unable to create scan result file %s: %s", lastScanFile, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.WriteString(tagText)
	if err != nil {
		return fmt.Errorf("unable to write scan result file %s: %s", lastScanFile, err)
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

	logger := service.NewLogger(appName)

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
			return startService(logger, cfg)
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
