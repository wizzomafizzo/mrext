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
// TODO: play a fun sound when a scan is successful or fails
// TODO: a -test command to see what the result of an NDEF would be
// TODO: would it be possible to unlock the OSD with a card?
// TODO: more concrete amiibo support
// TODO: cache the nfc db in memory and reload on inotify change
// TODO: strip colons from UID mapping file entries and make lowercase
// TODO: create a test web nfc reader in separate github repo, hosted on pages
// TODO: way to check the status of the service
// TODO: use a tag to signal that that next tag should have the active game written to it

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
	// TODO: move these to config with new names
	databaseFile = filepath.Join(config.SdFolder, "nfc.csv")
	lastScanFile = filepath.Join(config.TempFolder, "NFCSCAN")
)

type Card struct {
	CardType string
	UID      string
	Text     string
	ScanTime time.Time
}

func pollDevice(
	cfg *config.UserConfig,
	pnd *nfc.Device,
	lastSeen Card,
) (Card, error) {
	count, target, err := pnd.InitiatorPollTarget(supportedCardTypes, timesToPoll, periodBetweenPolls)
	if err != nil && !errors.Is(err, nfc.Error(nfc.ETIMEOUT)) {
		return lastSeen, fmt.Errorf("error polling: %s", err)
	}

	if count <= 0 {
		if lastSeen.UID != "" && time.Since(lastSeen.ScanTime) > timeToForgetCard {
			logger.Info("card removed")
			lastSeen = Card{}
		}

		return lastSeen, nil
	}

	cardUid := getCardUID(target)
	if cardUid == "" {
		logger.Warn("unable to detect card UID: %s", target.String())
	}

	if cardUid == lastSeen.UID {
		return lastSeen, nil
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
		return lastSeen, fmt.Errorf("error reading record: %s", err)
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

	err = writeScanResult(card.UID, card.Text)
	if err != nil {
		logger.Warn("error writing tmp scan result: %s", err)
	}

	err = launchCard(cfg, card)
	if err != nil {
		logger.Error("error launching card: %s", err)
	}

	return card, nil
}

func startService(cfg *config.UserConfig) (func() error, error) {
	var stopService bool
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

		lastSeen := Card{}

		for {
			if stopService {
				break
			}

			newSeen, err := pollDevice(cfg, &pnd, lastSeen)
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

func writeScanResult(uid string, text string) error {
	f, err := os.Create(lastScanFile)
	if err != nil {
		return fmt.Errorf("unable to create scan result file %s: %s", lastScanFile, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.WriteString(fmt.Sprintf("%s,%s", uid, text))
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
