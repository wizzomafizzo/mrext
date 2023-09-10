package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/clausecker/nfc/v2"
	"github.com/hsanjuan/go-ndef"
	"golang.org/x/exp/slices"
)

const (
	TypeNTAG                = "NTAG"
	TypeMifare              = "MIFARE"
	WRITE_COMMAND           = byte(0xA2)
	READ_COMMAND            = byte(0x30)
	NTAG_213_CAPACITY_BYTES = 114
	NTAG_215_CAPACITY_BYTES = 496
	NTAG_216_CAPACITY_BYTES = 872
)

var NDEF_END = []byte{0xFE}
var NDEF_START = []byte{0x54, 0x02, 0x65, 0x6E}

func readRecord(pnd nfc.Device, blockCount int) ([]byte, error) {
	allBlocks := make([]byte, 0)
	offset := 4

	for i := 0; i <= (blockCount / 4); i++ {
		blocks, err := comm(pnd, []byte{0x30, byte(offset)}, 16)
		if err != nil {
			return nil, err
		}
		allBlocks = append(allBlocks, blocks...)
		offset = offset + 4
	}

	return allBlocks, nil
}

func parseRecordText(blocks []byte) string {
	// Find the text NDEF record
	startIndex := bytes.Index(blocks, NDEF_START)
	endIndex := bytes.Index(blocks, NDEF_END)

	if startIndex != -1 && endIndex != -1 {
		tagText := string(blocks[startIndex+4 : endIndex])
		return tagText
	}

	return ""
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
		uid = ""
	}
	return uid
}

func comm(pnd nfc.Device, tx []byte, replySize int) ([]byte, error) {
	rx := make([]byte, replySize)

	timeout := 0
	_, err := pnd.InitiatorTransceiveBytes(tx, rx, timeout)
	if err != nil {
		return nil, fmt.Errorf("error reading block: %s", err)
	}

	return rx, nil
}

func getNtagCapacity(pnd nfc.Device) (int, error) {
	// Find tag capacity by looking in block 3 (capability container)
	tx := []byte{0x30, 0x03}
	rx := make([]byte, 16)

	timeout := 0
	_, err := pnd.InitiatorTransceiveBytes(tx, rx, timeout)
	if err != nil {
		return 0, err
	}

	switch rx[2] {
	case 0x12:
		// NTAG213. (144 -4) / 4
		return 35, nil
	case 0x3E:
		// NTAG215. (504 - 4) / 4
		return 125, nil
	case 0x6D:
		// NTAG216. (888 -4) / 4
		return 221, nil
	default:
		// fallback to NTAG213
		return 35, nil
	}
}

func getCardType(target nfc.Target) string {
	switch target.Modulation() {
	case nfc.Modulation{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106}:
		var card = target.(*nfc.ISO14443aTarget)
		if card.Atqa == [2]byte{0x00, 0x04} && card.Sak == 0x08 {
			// https://www.nxp.com/docs/en/application-note/AN10833.pdf page 9
			return TypeMifare
		}
		if card.Atqa == [2]byte{0x00, 0x44} && card.Sak == 0x00 {
			// https://www.nxp.com/docs/en/data-sheet/NTAG213_215_216.pdf page 33
			return TypeNTAG
		}
	}
	return ""
}

func authMifareCommand(block byte, cardUid string) []byte {
	command := []byte{
		// Auth using key A
		0x60, block,
		// Using the NDEF well known private key
		0xd3, 0xf7, 0xd3, 0xf7, 0xd3, 0xf7,
	}
	// And finally append the card UID to the end
	uidBytes, _ := hex.DecodeString(cardUid)
	return append(command, uidBytes...)

}

func readMifare(pnd nfc.Device, cardUid string) ([]byte, error) {

	permissionSectors := []int{4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60}
	var allBlocks = []byte{}
	for block := 0; block < 64; block++ {
		if block <= 3 {
			// The first sector contains infomation we don't care about and
			// also has a different key (0xA0A1A2A3A4A5) YAGNI, so skip over
			continue
		}

		// The last block of a sector contains KeyA + Permissions + KeyB
		// We don't care about that info so skip if present.
		if slices.Contains(permissionSectors, block+1) {
			continue
		}

		// Mifare is split up into 16 sectors each containing 4 blocks.
		// We need to authenticate before any read/ write operations can be performed
		// Only need to authenticate once per sector
		if block%4 == 0 {
			comm(pnd, authMifareCommand(byte(block), cardUid), 2)
		}

		blockData, err := comm(pnd, []byte{0x30, byte(block)}, 16)
		if err != nil {
			return nil, err
		}

		allBlocks = append(allBlocks, blockData...)

		if bytes.Contains(blockData, NDEF_END) {
			// Once we find the end of the NDEF text record there is no need to
			// continue reading the rest of the card.
			// This should make things "load" quicker
			break
		}

	}
	return allBlocks, nil
}

func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

// Only supports NTAG.
// Mifare requires an authentication call and a different write method (0xA0)
func writeTextToCard(pnd nfc.Device, text string) ([]byte, error) {
	var payload, err = BuildMessage(text)
	if err != nil {
		return nil, err
	}

	cardCapacity, err := getNtagCapacity(pnd)
	if err != nil {
		return nil, err
	}

	if len(payload) > cardCapacity {
		return nil, errors.New(fmt.Sprintf("Payload too big for card: [%d/%d] bytes used\n", len(payload), cardCapacity))
	}

	var startingBlock byte = 0x04
	for i, chunk := range chunkBy(payload, 4) {
		for len(chunk) < 4 {
			chunk = append(chunk, []byte{0x00}...)
		}
		var tx = []byte{WRITE_COMMAND, startingBlock + byte(i)}
		tx = append(tx, chunk...)
		_, err := comm(pnd, tx, 1)
		if err != nil {
			return nil, err
		}
	}

	return payload, nil
}

func getNtagCapacity(pnd nfc.Device) (int, error) {
	// Find tag capacity by looking in block 3 (capability container)
	tx := []byte{READ_COMMAND, 0x03}
	rx := make([]byte, 16)

	timeout := 0
	_, err := pnd.InitiatorTransceiveBytes(tx, rx, timeout)
	if err != nil {
		return 0, err
	}

	// https://github.com/adafruit/Adafruit_MFRC630/blob/master/docs/NTAG.md#capability-container
	switch rx[2] {
	case 0x12:
		return NTAG_213_CAPACITY_BYTES, nil
	case 0x3E:
		return NTAG_215_CAPACITY_BYTES, nil
	case 0x6D:
		return NTAG_216_CAPACITY_BYTES, nil
	default:
		// fallback
		return NTAG_213_CAPACITY_BYTES, nil
	}
}

func BuildMessage(message string) ([]byte, error) {
	ndef := ndef.NewTextMessage(message, "en")
	var payload, err = ndef.Marshal()
	if err != nil {
		return nil, err
	}

	var header, _ = CalculateNdefHeader(payload)
	if err != nil {
		return nil, err
	}
	payload = append(header, payload...)
	payload = append(payload, []byte{0xFE}...)
	return payload, nil
}

func CalculateNdefHeader(ndefRecord []byte) ([]byte, error) {
	var recordLength = len(ndefRecord)
	if recordLength < 255 {
		return []byte{0x03, byte(len(ndefRecord))}, nil
	}

	// NFCForum-TS-Type-2-Tag_1.1.pdf Page 9
	// > 255 Use three consecutive bytes format
	len := new(bytes.Buffer)
	err := binary.Write(len, binary.BigEndian, uint16(recordLength))
	if err != nil {
		return nil, err
	}

	var header = []byte{0x03, 0xFF}
	return append(header, len.Bytes()...), nil

}
