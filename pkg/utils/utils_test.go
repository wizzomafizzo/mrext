package utils

import (
	"os"
	"reflect"
	"testing"
)

func TestIsZip(t *testing.T) {
	var tests = []struct {
		path string
		want bool
	}{
		{"foo.zip", true},
		{"/path/to/foo.ZIP", true},
		{"foo.tar.gz", false},
		{"foo.txt", false},
		{"testdata/test.zip", true},
	}
	for _, tt := range tests {
		if got := IsZip(tt.path); got != tt.want {
			t.Errorf("IsZip(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestListZip(t *testing.T) {
	contents, err := ListZip("testdata/test.zip")
	if err != nil {
		t.Error(err)
	}
	if len(contents) != 5 {
		t.Errorf("got %d files, want 5", len(contents))
	}
	if contents[0] != "f1" {
		t.Errorf("got %q, want f1", contents[0])
	}
}

func TestMoveFile(t *testing.T) {
	orig_path := "testdata/test_file"
	dest_path := "testdata/test_file_moved"
	if _, err := os.Stat(orig_path); err != nil {
		t.Errorf("test file not found: %s", orig_path)
	}
	if err := MoveFile(orig_path, dest_path); err != nil {
		t.Errorf("error moving orig file: %s", err)
	}
	if _, err := os.Stat(orig_path); err == nil {
		t.Errorf("orig file still exists: %s", orig_path)
	}
	if _, err := os.Stat(dest_path); err != nil {
		t.Errorf("dest file not found: %s", dest_path)
	}
	if err := MoveFile(dest_path, orig_path); err != nil {
		t.Errorf("error moving dest file: %s", err)
	}
	if _, err := os.Stat(dest_path); err == nil {
		t.Errorf("dest file still exists: %s", dest_path)
	}
	if _, err := os.Stat(orig_path); err != nil {
		t.Errorf("orig file not found: %s", orig_path)
	}
}

func TestMax(t *testing.T) {
	var tests = []struct {
		xs   []int
		want int
	}{
		{[]int{}, 0},
		{[]int{1}, 1},
		{[]int{1, 2}, 2},
		{[]int{2, 1}, 2},
		{[]int{1, 2, 3}, 3},
		{[]int{3, 2, 1}, 3},
		{[]int{1, 2, 3, 4}, 4},
		{[]int{4, 2, 1, 3}, 4},
	}
	for _, tt := range tests {
		if got := Max(tt.xs); got != tt.want {
			t.Errorf("Max(%v) = %v, want %v", tt.xs, got, tt.want)
		}
	}
}

func TestMin(t *testing.T) {
	var tests = []struct {
		xs   []int
		want int
	}{
		{[]int{}, 0},
		{[]int{1}, 1},
		{[]int{1, 2}, 1},
		{[]int{2, 1}, 1},
		{[]int{1, 2, 3}, 1},
		{[]int{3, 2, 1}, 1},
		{[]int{1, 2, 3, 4}, 1},
		{[]int{4, 2, 1, 3}, 1},
	}
	for _, tt := range tests {
		if got := Min(tt.xs); got != tt.want {
			t.Errorf("Min(%v) = %v, want %v", tt.xs, got, tt.want)
		}
	}
}

func TestContains(t *testing.T) {
	var tests = []struct {
		xs   []int
		x    int
		want bool
	}{
		{[]int{}, 0, false},
		{[]int{1}, 1, true},
		{[]int{1, 2}, 1, true},
		{[]int{2, 1}, 3, false},
	}
	for _, tt := range tests {
		if got := Contains(tt.xs, tt.x); got != tt.want {
			t.Errorf("Contains(%v, %d) = %v, want %v", tt.xs, tt.x, got, tt.want)
		}
	}
}

func TestRandomElem(t *testing.T) {
	t1 := []string{}
	if el, err := RandomElem(t1); err == nil {
		t.Errorf("RandomElem(%v) = %q, want error", t1, el)
	}
	t2 := []int{1, 2, 3}
	el, err := RandomElem(t2)
	if err != nil {
		t.Errorf("RandomElem(%v) = %q, want no error", t2, el)
	}
	if !Contains(t2, el) {
		t.Errorf("RandomElem(%v) = %q, want element in %v", t2, el, t2)
	}
}

func TestMapKeys(t *testing.T) {
	var tests = []struct {
		m    map[string]int
		want []string
	}{
		{map[string]int{}, []string{}},
		{map[string]int{"a": 1}, []string{"a"}},
		{map[string]int{"a": 1, "b": 2}, []string{"a", "b"}},
	}
	for _, tt := range tests {
		if got := MapKeys(tt.m); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("MapKeys(%v) = %v, want %v", tt.m, got, tt.want)
		}
	}
}
