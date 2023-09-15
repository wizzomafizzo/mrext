package main

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

func TestCalculateNdefHeader(t *testing.T) {
	test2 := map[string]struct {
		input []byte
		want  []byte
	}{
		"minimum": {input: bytes.Repeat([]byte{0x69}, 1), want: []byte{0x03, 0x01}},
		"255":     {input: bytes.Repeat([]byte{0x69}, 255), want: []byte{0x03, 0xFF, 0x00, 0xFF}},
		"256":     {input: bytes.Repeat([]byte{0x69}, 256), want: []byte{0x03, 0xFF, 0x01, 0x00}},
		"257":     {input: bytes.Repeat([]byte{0x69}, 257), want: []byte{0x03, 0xFF, 0x01, 0x01}},
		"258":     {input: bytes.Repeat([]byte{0x69}, 258), want: []byte{0x03, 0xFF, 0x01, 0x02}},
		"512":     {input: bytes.Repeat([]byte{0x69}, 512), want: []byte{0x03, 0xFF, 0x02, 0x00}},
		"maximum": {input: bytes.Repeat([]byte{0x69}, 865), want: []byte{0x03, 0xFF, 0x03, 0x61}},
	}

	for name, tc := range test2 {
		t.Run(name, func(t *testing.T) {
			got, err := CalculateNdefHeader(tc.input)
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("test %v, expected: %v, got: %v", name, hex.EncodeToString(tc.want), hex.EncodeToString(got))
			}
		})

	}
}

func TestBuildMessage(t *testing.T) {
	test2 := []struct {
		input string
		want  string
	}{
		{input: "**random:snes", want: "0314d101105402656e2a2a72616e646f6d3a736e6573fe"},
		{input: "A", want: "0308d101045402656e41fe"},
		{input: "AAAA", want: "030bd101075402656e41414141fe"},
		{input: strings.Repeat("A", 512), want: "03ff020ac101000002035402656e4141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141fe"},
	}

	for name, tc := range test2 {
		t.Run(tc.input, func(t *testing.T) {
			got, err := BuildMessage(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			want, err := hex.DecodeString(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("test %v, expected: %v, got: %v", name, hex.EncodeToString(want), hex.EncodeToString(got))
			}
		})
	}
}
