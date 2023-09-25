package main

import (
	"encoding/hex"
	"fmt"

	"github.com/clausecker/nfc/v2"
)

const (
	TypeNTAG      = "NTAG"
	TypeMifare    = "MIFARE"
	WRITE_COMMAND = byte(0xA2)
	READ_COMMAND  = byte(0x30)
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
		return nil, fmt.Errorf("comm error: %s", err)
	}

	return rx, nil
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
