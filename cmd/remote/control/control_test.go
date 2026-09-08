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

package control

import (
	"slices"
	"testing"

	"github.com/bendahl/uinput"
	"github.com/wizzomafizzo/mrext/pkg/input"
)

// keyEvent is one uinput key transition. Recording both edges is what makes
// a combo distinguishable from consecutive presses of the same keys.
type keyEvent struct {
	code int
	down bool
}

// fakeKeyboard records what a real virtual keyboard would emit, so the named
// key table can be asserted without a uinput device.
type fakeKeyboard struct {
	events []keyEvent
}

func (f *fakeKeyboard) KeyPress(code int) error {
	f.events = append(f.events, keyEvent{code, true}, keyEvent{code, false})
	return nil
}

func (f *fakeKeyboard) KeyDown(code int) error {
	f.events = append(f.events, keyEvent{code, true})
	return nil
}

func (f *fakeKeyboard) KeyUp(code int) error {
	f.events = append(f.events, keyEvent{code, false})
	return nil
}

func (*fakeKeyboard) FetchSyspath() (string, error) {
	return "", nil
}

func (*fakeKeyboard) Close() error {
	return nil
}

// press is the event pair produced by input.Keyboard.Press.
func press(code int) []keyEvent {
	return []keyEvent{{code, true}, {code, false}}
}

// combo is the event sequence produced by input.Keyboard.Combo: every key is
// held in order, then released in that same order.
func combo(codes ...int) []keyEvent {
	events := make([]keyEvent, 0, len(codes)*2)
	for _, code := range codes {
		events = append(events, keyEvent{code, true})
	}
	for _, code := range codes {
		events = append(events, keyEvent{code, false})
	}
	return events
}

// TestSendKeyboardNamedKeys pins every name accepted by SendKeyboard to the
// keys it sends. The names are a published part of the Remote API, so a
// changed mapping here is a change to what existing clients do to a MiSTer.
func TestSendKeyboardNamedKeys(t *testing.T) {
	tests := []struct {
		name string
		want []keyEvent
	}{
		{"up", press(uinput.KeyUp)},
		{"down", press(uinput.KeyDown)},
		{"left", press(uinput.KeyLeft)},
		{"right", press(uinput.KeyRight)},
		{"volume_up", press(uinput.KeyVolumeup)},
		{"volume_down", press(uinput.KeyVolumedown)},
		{"volume_mute", press(uinput.KeyMute)},
		{"menu", press(uinput.KeyEsc)},
		{"back", press(uinput.KeyBackspace)},
		{"confirm", press(uinput.KeyEnter)},
		{"cancel", press(uinput.KeyEsc)},
		{"osd", press(uinput.KeyF12)},
		{"screenshot", combo(uinput.KeyLeftalt, uinput.KeyScrolllock)},
		{"raw_screenshot", combo(uinput.KeyLeftalt, uinput.KeyLeftshift, uinput.KeyScrolllock)},
		{"pair_bluetooth", press(uinput.KeyF11)},
		{"change_background", press(uinput.KeyF1)},
		{"core_select", combo(uinput.KeyLeftalt, uinput.KeyF12)},
		{"user", combo(uinput.KeyLeftctrl, uinput.KeyLeftalt, uinput.KeyRightalt)},
		{"reset", combo(uinput.KeyLeftshift, uinput.KeyLeftctrl, uinput.KeyLeftalt, uinput.KeyRightalt)},
		{"toggle_core_dates", press(uinput.KeyF2)},
		{"console", press(uinput.KeyF9)},
		{"exit_console", press(uinput.KeyF12)},
		{"computer_osd", combo(uinput.KeyLeftmeta, uinput.KeyF12)},
		{"save_state", combo(uinput.KeyLeftalt, uinput.KeyF1)},
		{"save_state_1", combo(uinput.KeyLeftalt, uinput.KeyF1)},
		{"save_state_2", combo(uinput.KeyLeftalt, uinput.KeyF2)},
		{"save_state_3", combo(uinput.KeyLeftalt, uinput.KeyF3)},
		{"save_state_4", combo(uinput.KeyLeftalt, uinput.KeyF4)},
		{"load_state", press(uinput.KeyF1)},
		{"load_state_1", press(uinput.KeyF1)},
		{"load_state_2", press(uinput.KeyF2)},
		{"load_state_3", press(uinput.KeyF3)},
		{"load_state_4", press(uinput.KeyF4)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeKeyboard{}
			kbd := input.Keyboard{Device: fake}

			if err := SendKeyboard(kbd, tt.name); err != nil {
				t.Fatalf("send %s: %v", tt.name, err)
			}

			if !slices.Equal(fake.events, tt.want) {
				t.Errorf("%s sent %v, want %v", tt.name, fake.events, tt.want)
			}
		})
	}
}

// TestSendKeyboardUnknownKey checks an unrecognised name is rejected rather
// than silently doing nothing, which is what makes the handler return 500.
func TestSendKeyboardUnknownKey(t *testing.T) {
	fake := &fakeKeyboard{}
	kbd := input.Keyboard{Device: fake}

	if err := SendKeyboard(kbd, "not_a_key"); err == nil {
		t.Fatal("expected an error for an unknown key name")
	}

	if len(fake.events) != 0 {
		t.Errorf("unknown key sent %v, want nothing", fake.events)
	}
}
