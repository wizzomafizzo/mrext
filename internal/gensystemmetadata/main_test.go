package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseAliases(t *testing.T) {
	t.Parallel()

	source := []byte(`package systemdefs
const SystemGenesis = "Genesis"
var Systems = map[string]System{
	SystemGenesis: {
		ID: SystemGenesis,
		Aliases: []string{"MegaDrive", "Mega Drive"},
	},
}
`)
	aliases, err := parseAliases(source)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{"Genesis": {"MegaDrive", "Mega Drive"}}
	if !reflect.DeepEqual(aliases, want) {
		t.Fatalf("aliases = %#v, want %#v", aliases, want)
	}
}

func TestGeneratedOutputCurrent(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "metadata.json")
	data := []byte(`{"source":"` + pinnedSourceLabel() + `","format":1,"systems":{}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if !generatedOutputCurrent(path) {
		t.Fatal("current generated output was not recognized")
	}
}
