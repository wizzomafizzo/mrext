package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/clausecker/nfc/v2"
	"golang.org/x/exp/slices"
)

const (
	MIFARE_WRITABLE_SECTOR_COUNT      = 15
	MIFARE_WRITABLE_BLOCKS_PER_SECTOR = 3
	MIFARE_BLOCK_SIZE_BYTES           = 16
)

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

func getMifareCapacityInBytes() int {
	return (MIFARE_WRITABLE_BLOCKS_PER_SECTOR * MIFARE_WRITABLE_SECTOR_COUNT) * MIFARE_BLOCK_SIZE_BYTES
}

func writeMifare(pnd nfc.Device, text string, cardUid string) ([]byte, error) {
	var payload, err = BuildMessage(text)
	if err != nil {
		return nil, err
	}

	var cardCapacity = getMifareCapacityInBytes()
	if len(payload) > cardCapacity {
		return nil, errors.New(fmt.Sprintf("Payload too big for card: [%d/%d] bytes used\n", len(payload), cardCapacity))
	}

	chunks := [][]byte{}
	for _, chunk := range chunkBy(payload, 16) {
		for len(chunk) < 16 {
			chunk = append(chunk, []byte{0x00}...)
		}
		chunks = append(chunks, chunk)
	}

	var chunkIndex = 0
	for sector := 1; sector <= 15; sector++ {
		// Iterate over blocks in sector (0-2) skipping trailer block (3)
		for sectorIndex := 0; sectorIndex < 3; sectorIndex++ {
			blockToWrite := (sector * 4) + sectorIndex
			if sectorIndex == 0 {
				// We changed sectors, time to authenticate
				_, err := comm(pnd, authMifareCommand(byte(blockToWrite), cardUid), 2)
				if err != nil {
					return nil, err
				}
			}

			writeBlockCommand := append([]byte{0xA0, byte(blockToWrite)}, chunks[chunkIndex]...)
			_, err := comm(pnd, writeBlockCommand, 2)
			if err != nil {
				return nil, err
			}
			chunkIndex++
			if chunkIndex >= len(chunks) {
				// All data has been written, we are done
				return payload, nil
			}
		}
	}

	return payload, nil
}
