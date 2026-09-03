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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bendahl/uinput"
)

const (
	doubleClickDelay = 100 * time.Millisecond
	clickHoldTime    = 60 * time.Millisecond
	absMax           = 1023
)

// Mouse owns virtual relative-mouse and absolute-touchpad devices.
type Mouse struct {
	Rel uinput.Mouse
	Abs uinput.TouchPad
	mu  *sync.Mutex
}

func NewMouse() (Mouse, error) {
	var mouse Mouse

	relative, err := uinput.CreateMouse("/dev/uinput", []byte("mrext-mouse"))
	if err != nil {
		return mouse, fmt.Errorf("create relative mouse: %w", err)
	}

	absolute, err := uinput.CreateTouchPad(
		"/dev/uinput", []byte("mrext-touchpad"),
		0, absMax, 0, absMax,
	)
	if err != nil {
		_ = relative.Close()
		return mouse, fmt.Errorf("create absolute touchpad: %w", err)
	}

	mouse.Rel = relative
	mouse.Abs = absolute
	mouse.mu = &sync.Mutex{}
	return mouse, nil
}

func (m *Mouse) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return errors.Join(m.Rel.Close(), m.Abs.Close())
}

func (m *Mouse) Move(x, y int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Rel.Move(x, y); err != nil {
		return fmt.Errorf("move relative mouse: %w", err)
	}
	return nil
}

// MoveToPermille moves cursor to per-mille coordinates from 0 through 1000.
func (m *Mouse) MoveToPermille(x, y int) error {
	clamp := func(value int) int {
		if value < 0 {
			return 0
		}
		if value > 1000 {
			return 1000
		}
		return value
	}

	// #nosec G115 -- clamped values produce results from zero through absMax.
	absoluteX := int32(clamp(x) * absMax / 1000)
	// #nosec G115 -- clamped values produce results from zero through absMax.
	absoluteY := int32(clamp(y) * absMax / 1000)

	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Abs.MoveTo(absoluteX, absoluteY); err != nil {
		return fmt.Errorf("move absolute mouse: %w", err)
	}
	return nil
}

func click(press, release func() error) error {
	if err := press(); err != nil {
		return fmt.Errorf("press mouse button: %w", err)
	}
	time.Sleep(clickHoldTime)
	if err := release(); err != nil {
		return fmt.Errorf("release mouse button: %w", err)
	}
	return nil
}

func (m *Mouse) clickLeftUnlocked() error {
	return click(m.Rel.LeftPress, m.Rel.LeftRelease)
}

func (m *Mouse) LeftClick() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.clickLeftUnlocked()
}

func (m *Mouse) DoubleClick() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.clickLeftUnlocked(); err != nil {
		return err
	}
	time.Sleep(doubleClickDelay)
	return m.clickLeftUnlocked()
}

func (m *Mouse) RightClick() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return click(m.Rel.RightPress, m.Rel.RightRelease)
}

func (m *Mouse) MiddleClick() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return click(m.Rel.MiddlePress, m.Rel.MiddleRelease)
}

func (m *Mouse) LeftDown() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Rel.LeftPress(); err != nil {
		return fmt.Errorf("press left mouse button: %w", err)
	}
	return nil
}

func (m *Mouse) LeftUp() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Rel.LeftRelease(); err != nil {
		return fmt.Errorf("release left mouse button: %w", err)
	}
	return nil
}

func (m *Mouse) RightDown() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Rel.RightPress(); err != nil {
		return fmt.Errorf("press right mouse button: %w", err)
	}
	return nil
}

func (m *Mouse) RightUp() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Rel.RightRelease(); err != nil {
		return fmt.Errorf("release right mouse button: %w", err)
	}
	return nil
}

func (m *Mouse) MiddleDown() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Rel.MiddlePress(); err != nil {
		return fmt.Errorf("press middle mouse button: %w", err)
	}
	return nil
}

func (m *Mouse) MiddleUp() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Rel.MiddleRelease(); err != nil {
		return fmt.Errorf("release middle mouse button: %w", err)
	}
	return nil
}

func (m *Mouse) Wheel(delta int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.Rel.Wheel(false, delta); err != nil {
		return fmt.Errorf("move mouse wheel: %w", err)
	}
	return nil
}
