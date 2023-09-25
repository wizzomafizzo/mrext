package main

import (
	"errors"
	"fmt"

	"github.com/clausecker/nfc/v2"
)

const (
	NTAG_213_CAPACITY_BYTES = 114
	NTAG_215_CAPACITY_BYTES = 496
	NTAG_216_CAPACITY_BYTES = 872
)

// Only supports NTAG.
// Mifare requires an authentication call and a different write method (0xA0)
func writeNtag(pnd nfc.Device, text string) ([]byte, error) {
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

func getNtagBlockCount(pnd nfc.Device) (int, error) {
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

func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}
