package main

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	nfc "github.com/clausecker/nfc/v2"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

var (
	nfcConnectionString = "" // Use the first NFC reader available. Be sure to configure /etc/nfc/libnfc.conf
	supportedCardTypes  = []nfc.Modulation{
		{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106},
	}
	timesToPoll        = 20
	periodBetweenPolls = 300 * time.Millisecond
	periodBetweenLoop  = 300 * time.Millisecond
	database           = make(map[string]string)
	databaseFile       = "/media/fat/nfc-mapping.csv"
)

func main() {
	log.Println("MiSTer NFC Reader (libnfc version" + nfc.Version() + ")")

	loadDatabase()

	pnd, err := nfc.Open(nfcConnectionString)
	if err != nil {
		log.Fatalln("Could not open device: ", err)
	}
	defer pnd.Close()

	if err := pnd.InitiatorInit(); err != nil {
		log.Fatalln("Could not init initiator: ", err)
	}

	log.Println("Opened: ", pnd, pnd.Connection())

	var lastSeenCardUID = ""

	for {
		var count, target, error = pnd.InitiatorPollTarget(supportedCardTypes, timesToPoll, periodBetweenPolls)

		if error != nil {
			log.Fatalln("Error polling: ", error)
		}

		if count > 0 {
			var currentCardID = getCardUID(target)
			if currentCardID != lastSeenCardUID {
				log.Println("New card UID: " + currentCardID)
				lastSeenCardUID = currentCardID
				var tagText = readTextRecord(pnd)
				if tagText != "" {
					log.Printf("Decoded text NDEF is: %s\n", tagText)
					loadCoreFromFilename(tagText)
					// TODO: if string is in special format e.g. !!GBA:abad9c764c35b8202e3d9e5915ca7007bdc7cc62 try to load that way.
				} else {
					log.Printf("No text NDEF found, falling back to UID mapping in CSV file")
					loadCoreFromCardUID(currentCardID)
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
	log.Printf("Card hex: " + hex.EncodeToString(allBlocks))

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
	data := readCsvFile(databaseFile)
	for _, row := range data {
		uid := row[0]
		value := row[1]
		database[uid] = value
	}
	log.Println("Loaded " + fmt.Sprint(len(database)) + " NFC mappings from the CSV")
}

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("Unable to load fallback database file: "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("CSV file appears to be badly formatted: "+filePath, err)
	}

	return records
}

func loadCoreFromFilename(filename string) {
	var fullpath = "/media/fat/" + filename // TODO: saves a few chars on the tag but is it worth it?
	if _, err := os.Stat(fullpath); errors.Is(err, os.ErrNotExist) {
		log.Println("Core does not exist: " + fullpath)
		return
	}

	log.Println("Loading core: " + fullpath)
	mister.LaunchGenericFile(fullpath)
}

func loadCoreFromCardUID(cardId string) {
	filename, ok := database[cardId]
	if !ok {
		log.Println("No core configured")
		return
	}

	loadCoreFromFilename(filename)
}

func getCardUID(target nfc.Target) string {
	var UID = ""
	switch target.Modulation() {
	case nfc.Modulation{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106}:
		var card = target.(*nfc.ISO14443aTarget)
		var ID = card.UID
		UID = hex.EncodeToString(ID[:card.UIDLen])
		break
	default:
		log.Println("Unsupported card type :(")
	}
	return UID
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
