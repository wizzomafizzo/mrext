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

package mister

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func mapSharedMem(address int64) (*[]byte, *os.File, error) {
	file, err := os.OpenFile(
		"/dev/mem",
		os.O_RDWR|os.O_SYNC,
		0,
	)
	if err != nil {
		return &[]byte{}, nil, fmt.Errorf("error opening /dev/mem: %w", err)
	}

	mem, err := syscall.Mmap(
		int(file.Fd()),
		address,
		0x1000,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		_ = file.Close()
		return &[]byte{}, nil, fmt.Errorf("error mapping /dev/mem: %w", err)
	}

	return &mem, file, nil
}

func unmapSharedMem(mem *[]byte, file *os.File) error {
	err := syscall.Munmap(*mem)
	if err != nil {
		return fmt.Errorf("error unmapping /dev/mem: %w", err)
	}

	if file == nil {
		return errors.New("/dev/mem file reference is nil")
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("error closing /dev/mem: %w", err)
	}

	return nil
}

func GetActiveIni() (int, error) {
	mem, file, err := mapSharedMem(0x1FFFF000)
	if err != nil {
		return 0, err
	}

	offset := 0xF04
	vs := []byte{(*mem)[offset], (*mem)[offset+1], (*mem)[offset+2], (*mem)[offset+3]}

	err = unmapSharedMem(mem, file)
	if err != nil {
		return 0, err
	}

	if vs[0] == 0x34 && vs[1] == 0x99 && vs[2] == 0xBA {
		return int(vs[3] + 1), nil
	}
	return 0, nil
}

func SetActiveIni(ini int, relaunchCore bool) error {
	if ini < 1 || ini > 4 {
		return fmt.Errorf("ini number out of range: %d", ini)
	}

	mem, file, err := mapSharedMem(0x1FFFF000)
	if err != nil {
		return err
	}

	offset := 0xF04
	(*mem)[offset] = 0x34
	(*mem)[offset+1] = 0x99
	(*mem)[offset+2] = 0xBA
	(*mem)[offset+3] = byte(ini - 1)

	err = unmapSharedMem(mem, file)
	if err != nil {
		return err
	}

	if !relaunchCore {
		return nil
	}

	coreName, err := GetActiveCoreName()
	if err != nil {
		return err
	}

	if coreName == config.MenuCore {
		return LaunchMenu()
	}

	// TODO: can we check if this file has been modified recently?
	recent, err := ReadRecent(config.CoresRecentFile)
	if err != nil || len(recent) == 0 {
		return LaunchMenu()
	}

	corePath := filepath.Join(config.SdFolder, recent[0].Directory, recent[0].Name)
	// TODO: use real config later
	return LaunchGenericFile(&config.UserConfig{}, corePath)
}
