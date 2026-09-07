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

package main

import (
	"errors"
	"net"
	"strconv"
	"syscall"
	"testing"
)

func TestBindFailureReturnsBeforeDeviceSetup(t *testing.T) {
	// No logger/config/devices are needed if bind fails. An injected listener
	// avoids touching either a real port or the MiSTer input devices.
	calls := 0
	stop, err := startServiceWithListener(nil, nil, func(network, address string) (net.Listener, error) {
		calls++
		if network != "tcp" || address != ":"+strconv.Itoa(appPort) {
			t.Fatalf("unexpected listener: %s %s", network, address)
		}
		return nil, syscall.EADDRINUSE
	})
	if stop != nil || !errors.Is(err, syscall.EADDRINUSE) || calls != 1 {
		t.Fatalf("bind failure hidden: stop=%v err=%v calls=%d", stop != nil, err, calls)
	}
}
