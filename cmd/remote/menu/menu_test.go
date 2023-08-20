package menu

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestLoadNamesMapping(t *testing.T) {
	t.Cleanup(resetState)

	err := loadNamesMapping(mockFilesystem())

	if err != nil {
		t.Errorf("Error loading names.txt: %v", err)
	}

	expectedSize := 2
	actualSize := len(namesMapping)
	if actualSize != expectedSize {
		t.Errorf("Incorrect number of names loaded: %d expected %d", actualSize, expectedSize)
	}

	scenarios := map[string]string{
		"Genesis":        "Mega Drive",
		"Genesis3D":      "Mega Drive 3D",
		"Unknown System": "",
	}

	for original, expected := range scenarios {
		t.Run(original, func(t *testing.T) {
			actual := namesMapping[original]
			if actual != expected {
				t.Errorf("Name not mapped correctly, expected %s got %s", expected, actual)
			}
		})
	}
}
func TestGetNamesTxtReturnsCorrectName(t *testing.T) {
	t.Cleanup(resetState)
	err := loadNamesMapping(mockFilesystem())
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := "Mega Drive"
	actual, err := GetNamesTxt("Genesis", "")
	if err != nil {
		t.Errorf("%v", err)
	}

	if actual != expected {
		t.Errorf("expected %s got %s", expected, actual)
	}
}
func TestGetNamesTxtReturnsEmptyStringForFolder(t *testing.T) {
	t.Cleanup(resetState)
	err := loadNamesMapping(mockFilesystem())
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := ""
	actual, err := GetNamesTxt("_Favourites", "folder")
	if err != nil {
		t.Errorf("%v", err)
	}

	if actual != expected {
		t.Errorf("expected %s got %s", expected, actual)
	}
}

func mockFilesystem() fs.FS {
	tfs := fstest.MapFS{"media/fat/names.txt": &fstest.MapFile{
		Data: []byte("Genesis:            Mega Drive\nGenesis3D:          Mega Drive 3D"),
	}}

	return tfs
}

func resetState() {
	namesMapping = map[string]string{}
}
