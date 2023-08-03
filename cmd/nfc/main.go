package main

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"

	"github.com/clausecker/nfc/v2"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

var (
	appName            = "nfc"
	supportedCardTypes = []nfc.Modulation{
		{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106},
	}
	timesToPoll        = 20
	periodBetweenPolls = 300 * time.Millisecond
	periodBetweenLoop  = 300 * time.Millisecond
	// TODO: i think this can be moved to main instead of being a global
	database     = make(map[string]string)
	databaseFile = "/media/fat/nfc-mapping.csv"
)

var logger = service.NewLogger(appName)

func main() {
	logger.Info("MiSTer NFC Reader (libnfc version %s)", nfc.Version())

	loadDatabase()

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{})
	if err != nil {
		logger.Error("error loading user config: %s", err)
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	pnd, err := nfc.Open(cfg.NfcConfig.ConnectionString)
	if err != nil {
		logger.Error("could not open device: %s", err)
		fmt.Println("Could not connect to NFC device:", err)
		os.Exit(1)
	}
	defer func(pnd nfc.Device) {
		err := pnd.Close()
		if err != nil {
			logger.Warn("error closing device: %s", err)
		}
	}(pnd)

	if err := pnd.InitiatorInit(); err != nil {
		logger.Error("could not init initiator: %s", err)
		fmt.Println("Could not initialize NFC device:", err)
		os.Exit(1)
	}

	logger.Info("opened connection: %s %s", pnd, pnd.Connection())

	var lastSeenCardUID string

	for {
		count, target, err := pnd.InitiatorPollTarget(supportedCardTypes, timesToPoll, periodBetweenPolls)
		if err != nil {
			logger.Error("error polling: %s", err)
			fmt.Println("Lost connection to NFC device:", err)
			//os.Exit(1)
		}

		if count > 0 {
			currentCardID := getCardUID(target)

			if currentCardID != lastSeenCardUID {
				logger.Info("new card UID: %s", currentCardID)
				lastSeenCardUID = currentCardID
				tagText := readTextRecord(pnd)

				if tagText != "" {
					logger.Info("decoded text NDEF: %s", tagText)
					loadCoreFromFilename(tagText)
					// TODO: if string is in special format
					//       e.g. !!GBA:abad9c764c35b8202e3d9e5915ca7007bdc7cc62 try to load that way.
				} else {
					logger.Info("no text NDEF found, falling back to UID mapping in CSV file")
					loadCoreFromCardUID(currentCardID)
					// TODO: check if this failed too and log
				}
			}
		}

		time.Sleep(periodBetweenLoop)
	}
}

func readTextRecord(pnd nfc.Device) string {
	blockCount := 35 // TODO: This is hardcoded for NTAG 213. needs to support N215 and N216
	allBlocks := make([]byte, 0)
	offset := 4
	for i := 0; i <= (blockCount / 4); i++ {
		blocks := readFourBlocks(pnd, byte(offset))
		allBlocks = append(allBlocks, blocks...)
		offset = offset + 4
	}
	logger.Info("card hex: %s", hex.EncodeToString(allBlocks))

	// Find the text NDEF record
	startIndex := bytes.Index(allBlocks, []byte{0x54, 0x02, 0x65, 0x6E})
	endIndex := bytes.Index(allBlocks, []byte{0xFE})

	if startIndex != -1 && endIndex != -1 {
		tagText := string(allBlocks[startIndex+4 : endIndex])
		return tagText
	}

	return "" // TODO: return error,string instead
}

func loadDatabase() {
	// TODO: need to return an error
	data := readCsvFile(databaseFile)
	for _, row := range data {
		uid := row[0]
		value := row[1]
		database[uid] = value
	}
	// TODO: return number of rows loaded and give friendly output in main
	logger.Info("loaded %d NFC mappings from the CSV", len(database))
}

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		logger.Error("unable to load fallback database file %s: %s", filePath, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logger.Warn("error closing file: %s", err)
		}
	}(f)

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		logger.Error("CSV file %s appears to be badly formatted: %s", filePath, err)
		// TODO: return an error and exit in main if necessary
		os.Exit(1)
	}

	return records
}

func loadCoreFromFilename(filename string) {
	fullPath := filepath.Join(config.SdFolder, filename) // TODO: saves a few chars on the tag but is it worth it?

	if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
		logger.Error("core does not exist: %s", fullPath)
		// TODO: return error
		return
	}
	logger.Info("loading core: %s", fullPath)

	// TODO: handle and return error
	_ = mister.LaunchGenericFile(fullPath)
}

func loadCoreFromCardUID(cardId string) {
	filename, ok := database[cardId]
	if !ok {
		logger.Error("no core mapped for card ID: %s", cardId)
		return
	}

	// TODO: return error?
	loadCoreFromFilename(filename)
}

func getCardUID(target nfc.Target) string {
	var uid string
	switch target.Modulation() {
	case nfc.Modulation{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106}:
		var card = target.(*nfc.ISO14443aTarget)
		var ID = card.UID
		uid = hex.EncodeToString(ID[:card.UIDLen])
		break
	default:
		// TODO: does target.String() give us anything useful?
		logger.Info("unsupported card type: %s", target.String())
	}
	return uid
}

func readFourBlocks(pnd nfc.Device, blockNumber byte) []byte {
	// Read 16 bytes at a time from a Type 2 tag
	// For NTAG this would be 4 blocks or pages.
	tx := []byte{0x30, blockNumber}
	rx := make([]byte, 16)

	timeout := 0
	_, err := pnd.InitiatorTransceiveBytes(tx, rx, timeout)
	if err != nil {
		fmt.Println("Error reading blocks: ", err)
		return nil
	}

	return rx
}
