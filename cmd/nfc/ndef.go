package main

import (
	"bytes"
	"encoding/binary"

	"github.com/hsanjuan/go-ndef"
)

func ParseRecordText(blocks []byte) string {
	// Find the text NDEF record
	startIndex := bytes.Index(blocks, NDEF_START)
	endIndex := bytes.Index(blocks, NDEF_END)

	if startIndex != -1 && endIndex != -1 {
		tagText := string(blocks[startIndex+4 : endIndex])
		return tagText
	}

	return ""
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
