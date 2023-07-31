package main

import (
	"encoding/csv"
	"encoding/hex"
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
		{Type: nfc.ISO14443b, BaudRate: nfc.Nbr106},
		{Type: nfc.Felica, BaudRate: nfc.Nbr212},
		{Type: nfc.Felica, BaudRate: nfc.Nbr424},
		{Type: nfc.Jewel, BaudRate: nfc.Nbr106},
		{Type: nfc.ISO14443biClass, BaudRate: nfc.Nbr106},
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
				loadCore(currentCardID)
			}
		}

		time.Sleep(periodBetweenLoop)
	}
}

func loadDatabase() {
	data := readCsvFile(databaseFile)
	for _, row := range data {
		uid := row[0]
		value := row[1]
		database[uid] = value
	}
	log.Println("Loaded " + fmt.Sprint(len(database)) + "NFC mappings from the CSV")
}

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

func loadCore(cardId string) {
	filename, ok := database[cardId]
	if !ok {
		log.Println("No core configured")
		return
	}

	log.Println("Loading core: " + filename)
	mister.LaunchGenericFile(filename)
}

func getCardUID(target nfc.Target) string {
	var UID = ""
	switch target.Modulation() {
	case nfc.Modulation{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106}:
		var card = target.(*nfc.ISO14443aTarget)
		var ID = card.UID
		UID = hex.EncodeToString(ID[:card.UIDLen])
		break
	case nfc.Modulation{Type: nfc.ISO14443b, BaudRate: nfc.Nbr106}:
		var card = target.(*nfc.ISO14443bTarget)
		UID = hex.EncodeToString(card.ApplicationData[:len(card.ApplicationData)])
		break
	case nfc.Modulation{Type: nfc.Felica, BaudRate: nfc.Nbr212}:
		var card = target.(*nfc.FelicaTarget)
		UID = hex.EncodeToString(card.ID[:card.Len])
		break
	case nfc.Modulation{Type: nfc.Felica, BaudRate: nfc.Nbr424}:
		var card = target.(*nfc.FelicaTarget)
		UID = hex.EncodeToString(card.ID[:card.Len])
		break
	case nfc.Modulation{Type: nfc.Jewel, BaudRate: nfc.Nbr106}:
		var card = target.(*nfc.JewelTarget)
		var ID = card.ID
		UID = hex.EncodeToString(ID[:len(ID)])
		break
	case nfc.Modulation{Type: nfc.ISO14443biClass, BaudRate: nfc.Nbr106}:
		var card = target.(*nfc.ISO14443biClassTarget)
		var ID = card.UID
		UID = hex.EncodeToString(ID[:len(ID)])
		break
	default:
		log.Println("Unsupported card type :(")
	}
	return UID
}
