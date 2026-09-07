// mrext
// Copyright (c) 2026 mrext contributors.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of mrext.
//
// mrext is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// mrext is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with mrext. If not, see <http://www.gnu.org/licenses/>.

package games

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadNeoGeoNames(t *testing.T) {
	names, err := ReadNeoGeoNames(strings.NewReader(`<romsets>
		<romset name="MSLUG, ms1, ," altname=" Metal &amp; Slug "/>
		<romset name="ignored" altname=" "/>
		<romset name="" altname="Ignored"/>
		<romset name="ms1" altname="Alias title"/>
	</romsets>`))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"mslug": "Metal & Slug", "ms1": "Alias title"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("names = %v, want %v", names, want)
	}
	names, err = ReadNeoGeoNames(strings.NewReader(`<romsets><romset name="partial" altname="Not valid"/>`))
	if err == nil || len(names) != 0 {
		t.Fatalf("malformed XML must not return partial names: %v, %v", names, err)
	}
}

func TestNeoGeoMGLOverrideUsesCatalogSlot(t *testing.T) {
	// Nonstandard parameters prove ZIP/folder support does not hardcode a
	// second copy of NeoGeo's catalog data.
	system := &System{Id: "NeoGeo", Slots: []Slot{{
		Exts: []string{".neo"},
		Mgl:  &MglParams{Delay: 3, Method: "s", Index: 2, resetDelay: 4, resetHold: 5},
	}}}
	override, err := NeoGeoMGLOverride(system, `Picks/A & B "Special" <set>.zip`)
	if err != nil {
		t.Fatal(err)
	}
	want := "\t<file delay=\"3\" type=\"s\" index=\"2\" " +
		"path=\"Picks/A &amp; B &quot;Special&quot; &lt;set&gt;.zip\"/>\n" +
		"\t<reset delay=\"4\" hold=\"5\"/>\n"
	if override != want {
		t.Fatalf("override = %q, want %q", override, want)
	}
	for _, invalid := range []*System{nil, {Id: "NES"}, {Id: "NeoGeo"}} {
		if _, err := NeoGeoMGLOverride(invalid, "set"); err == nil {
			t.Fatalf("expected invalid system or missing-slot error: %+v", invalid)
		}
	}
}

func FuzzReadNeoGeoNames(f *testing.F) {
	f.Add(`<romsets><romset name="mslug,ms1" altname="Metal Slug"/></romsets>`)
	f.Add(`<romsets><romset name="partial" altname="Title"/>`)
	f.Add("")
	f.Fuzz(func(t *testing.T, data string) {
		names, err := ReadNeoGeoNames(strings.NewReader(data))
		if err != nil && len(names) != 0 {
			t.Fatal("invalid XML returned a partial mapping")
		}
		for name, title := range names {
			if name == "" || name != strings.ToLower(strings.TrimSpace(name)) || strings.TrimSpace(title) == "" {
				t.Fatalf("invalid mapping: %q => %q", name, title)
			}
		}
	})
}
