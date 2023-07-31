package main

import (
	"encoding/hex"
	"log"
	"os"
	"time"

	nfc "github.com/clausecker/nfc/v2"
)

var (
	nfcConnectionString = ""             // Use the first NFC reader available. Be sure to configure /etc/nfc/libnfc.conf
	nfcIPCFile          = "/tmp/nfc-uid" // Where the card UID is written to
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
)

func main() {
	log.Println("MiSTer NFC Reader (libnfc version" + nfc.Version() + ")")

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
				err := os.WriteFile(nfcIPCFile, []byte(currentCardID), 0644)
				if err != nil {
					log.Fatalln("Unable to write card UID to filesystem: ", err)
				}
				lastSeenCardUID = currentCardID
			}
		}

		time.Sleep(periodBetweenLoop)
	}
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
