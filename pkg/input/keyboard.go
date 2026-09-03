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

package input

import (
	"fmt"
	"time"

	"github.com/bendahl/uinput"
)

// TODO: needs delays on connect if not running as a daemon

const sleepTime = 40 * time.Millisecond

type Keyboard struct {
	Device uinput.Keyboard
}

func NewKeyboard() (Keyboard, error) {
	var kb Keyboard

	vk, err := uinput.CreateKeyboard("/dev/uinput", []byte("mrext"))
	if err != nil {
		return kb, fmt.Errorf("create virtual keyboard: %w", err)
	}

	kb.Device = vk

	return kb, nil
}

func (k *Keyboard) Close() error {
	if err := k.Device.Close(); err != nil {
		return fmt.Errorf("close virtual keyboard: %w", err)
	}
	return nil
}

func (k *Keyboard) Press(key int) error {
	if err := k.KeyDown(key); err != nil {
		return err
	}
	time.Sleep(sleepTime)
	return k.KeyUp(key)
}

func (k *Keyboard) Combo(keys ...int) error {
	for _, key := range keys {
		if err := k.KeyDown(key); err != nil {
			return err
		}
	}
	time.Sleep(sleepTime)
	for _, key := range keys {
		if err := k.KeyUp(key); err != nil {
			return err
		}
	}
	return nil
}

func (k *Keyboard) KeyDown(key int) error {
	if err := k.Device.KeyDown(key); err != nil {
		return fmt.Errorf("press key %d: %w", key, err)
	}
	return nil
}

func (k *Keyboard) KeyUp(key int) error {
	if err := k.Device.KeyUp(key); err != nil {
		return fmt.Errorf("release key %d: %w", key, err)
	}
	return nil
}

func (k *Keyboard) VolumeUp() error {
	return k.Press(uinput.KeyVolumeup)
}

func (k *Keyboard) VolumeDown() error {
	return k.Press(uinput.KeyVolumedown)
}

func (k *Keyboard) VolumeMute() error {
	return k.Press(uinput.KeyMute)
}

func (k *Keyboard) Menu() error {
	return k.Press(uinput.KeyEsc)
}

func (k *Keyboard) Back() error {
	return k.Press(uinput.KeyBackspace)
}

func (k *Keyboard) Confirm() error {
	return k.Press(uinput.KeyEnter)
}

func (k *Keyboard) Cancel() error {
	return k.Menu()
}

func (k *Keyboard) Up() error {
	return k.Press(uinput.KeyUp)
}

func (k *Keyboard) Down() error {
	return k.Press(uinput.KeyDown)
}

func (k *Keyboard) Left() error {
	return k.Press(uinput.KeyLeft)
}

func (k *Keyboard) Right() error {
	return k.Press(uinput.KeyRight)
}

func (k *Keyboard) OSD() error {
	return k.Press(uinput.KeyF12)
}

func (k *Keyboard) CoreSelect() error {
	return k.Combo(uinput.KeyLeftalt, uinput.KeyF12)
}

func (k *Keyboard) Screenshot() error {
	// TODO: for the life of me, I can't make the regular Win+PrtScn combo
	//       work. this is a hardcoded alternate combo which *does* work,
	//       but it's disabled on PS/2 keyboard or in PS/2 mode or something
	return k.Combo(uinput.KeyLeftalt, uinput.KeyScrolllock)
}

func (k *Keyboard) RawScreenshot() error {
	// TODO: see above
	return k.Combo(uinput.KeyLeftalt, uinput.KeyLeftshift, uinput.KeyScrolllock)
}

func (k *Keyboard) User() error {
	return k.Combo(uinput.KeyLeftctrl, uinput.KeyLeftalt, uinput.KeyRightalt)
}

func (k *Keyboard) Reset() error {
	return k.Combo(uinput.KeyLeftshift, uinput.KeyLeftctrl, uinput.KeyLeftalt, uinput.KeyRightalt)
}

func (k *Keyboard) PairBluetooth() error {
	return k.Press(uinput.KeyF11)
}

func (k *Keyboard) ChangeBackground() error {
	return k.Press(uinput.KeyF1)
}

func (k *Keyboard) ToggleCoreDates() error {
	return k.Press(uinput.KeyF2)
}

func (k *Keyboard) Console() error {
	return k.Press(uinput.KeyF9)
}

func (k *Keyboard) ExitConsole() error {
	return k.Press(uinput.KeyF12)
}

func (k *Keyboard) ComputerOSD() error {
	return k.Combo(uinput.KeyLeftmeta, uinput.KeyF12)
}
