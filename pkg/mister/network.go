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
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/libp2p/zeroconf/v2"
	"github.com/txn2/txeh"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"golang.org/x/sys/unix"
)

const (
	DefaultHostname = "MiSTer"
	MDNSServiceName = "_mister-remote._tcp"
	mdnsPort        = 5353
	mdnsTTL         = 120
	startRetries    = 30
	discoveryTime   = 15 * time.Second
	browseInterval  = 1 * time.Minute
)

type MDNSClient struct {
	Hostname string
	Version  string
	IP       string
}

type MDNSService struct {
	Clients []MDNSClient
	mu      sync.Mutex
	Active  bool
}

func (s *MDNSService) AddClient(client MDNSClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Clients = append(s.Clients, client)
	s.Active = true
}

func (s *MDNSService) ClearClients() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Clients = []MDNSClient{}
}

func (s *MDNSService) GetClients() []MDNSClient {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]MDNSClient(nil), s.Clients...)
}

func (s *MDNSService) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Active
}

func (s *MDNSService) SetActive(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Active = active
}

var MDNS = &MDNSService{
	Active:  false,
	Clients: []MDNSClient{},
}

func browseMDNS(logger *service.Logger) {
	MDNS.ClearClients()

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			version := ""
			if len(entry.Text) > 0 {
				version = strings.Split(entry.Text[0], "=")[1]
			}

			ip := ""
			if len(entry.AddrIPv4) > 0 {
				ip = entry.AddrIPv4[0].String()
			}

			MDNS.AddClient(MDNSClient{
				Hostname: strings.TrimSuffix(entry.HostName, "."),
				Version:  version,
				IP:       ip,
			})
		}
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), discoveryTime)
	defer cancel()

	err := zeroconf.Browse(
		ctx,
		MDNSServiceName,
		"local.",
		entries,
		zeroconf.SelectIPTraffic(zeroconf.IPv4),
	)
	if err != nil {
		logger.Error("error during mdns browse: %s", err)
	}

	<-ctx.Done()
}

func startMDNS(logger *service.Logger, appVersion string) (func() error, error) {
	if MDNS.IsActive() {
		return func() error { return nil }, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("read hostname: %w", err)
	}

	server, err := zeroconf.Register(
		"MiSTer Remote ("+hostname+")",
		MDNSServiceName,
		"local.",
		mdnsPort,
		[]string{"version=" + appVersion},
		nil,
		zeroconf.TTL(mdnsTTL),
	)
	if err != nil {
		return nil, fmt.Errorf("register mDNS service: %w", err)
	}
	MDNS.SetActive(true)
	logger.Info("registered mdns service with hostname: %s", hostname)

	browseMDNS(logger)
	ticker := time.NewTicker(browseInterval)
	go func() {
		for range ticker.C {
			browseMDNS(logger)
		}
	}()
	logger.Info("started network discovery service")

	return func() error {
		ticker.Stop()
		server.Shutdown()
		MDNS.ClearClients()
		MDNS.SetActive(false)
		return nil
	}, nil
}

// TryStartMDNS will attempt to start the mDNS service, retrying multiple times if it fails. This is because a script
// may be run at boot time before the network is available.
func TryStartMDNS(logger *service.Logger, appVersion string) func() error {
	// TODO: allow a hook function on successful browse
	retries := 0
	for {
		stop, err := startMDNS(logger, appVersion)
		if err == nil {
			return stop
		}
		if retries >= startRetries {
			logger.Error("failed to start mdns service, giving up: %s", err)
			return nil
		}

		retries++
		if retries == 1 {
			logger.Error("failed to start mdns service, retrying: %s", err)
		}
		time.Sleep(time.Second)
	}
}

// UpdateHostname updates all hostname related files with the new hostname and refreshes it in kernel memory.
func UpdateHostname(newHostname string, writeProc bool) error {
	// TODO: also update the linux/hostname file and linux/hosts file
	procHostnameFile := "/proc/sys/kernel/hostname"
	hostnameFile := "/etc/hostname"
	localIP := "127.0.1.1"

	if newHostname == "" {
		newHostname = DefaultHostname
	}

	currentHostnameData, err := os.ReadFile(hostnameFile)
	if err != nil {
		return fmt.Errorf("read hostname file: %w", err)
	}

	currentHostname := string(currentHostnameData)

	if currentHostname == newHostname {
		// no change required
		return nil
	}

	if unix.Access("/", unix.W_OK) != nil {
		err = syscall.Mount("/", "/", "", syscall.MS_REMOUNT, "")
		if err != nil {
			return fmt.Errorf("remount root filesystem writable: %w", err)
		}

		defer func() {
			_ = syscall.Mount("/", "/", "", syscall.MS_REMOUNT|syscall.MS_RDONLY, "")
		}()
	}

	// update hostname file
	// #nosec G306 -- system hostname file must remain world-readable.
	err = os.WriteFile(hostnameFile, []byte(newHostname), 0o644)
	if err != nil {
		return fmt.Errorf("write hostname file: %w", err)
	}

	// update hosts file
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		return fmt.Errorf("load hosts file: %w", err)
	}

	hosts.RemoveHost(strings.ToLower(currentHostname))
	hosts.AddHost(localIP, strings.ToLower(newHostname))

	err = hosts.Save()
	if err != nil {
		return fmt.Errorf("save hosts file: %w", err)
	}

	// write new hostname to proc
	if writeProc {
		// #nosec G306 -- proc hostname node uses conventional readable permissions.
		err = os.WriteFile(procHostnameFile, []byte(newHostname), 0o644)
		if err != nil {
			return fmt.Errorf("update kernel hostname: %w", err)
		}
	}

	return nil
}

func FixRootSSHPerms() error {
	if unix.Access("/", unix.W_OK) != nil {
		if err := syscall.Mount("/", "/", "", syscall.MS_REMOUNT, ""); err != nil {
			return fmt.Errorf("remount root filesystem writable: %w", err)
		}
		defer func() {
			_ = syscall.Mount("/", "/", "", syscall.MS_REMOUNT|syscall.MS_RDONLY, "")
		}()
	}

	// #nosec G302 -- directories require execute permission for traversal.
	if err := os.Chmod(config.SSHConfigFolder, 0o700); err != nil {
		return fmt.Errorf("secure SSH configuration directory: %w", err)
	}

	root, err := os.OpenRoot(config.SSHConfigFolder)
	if err != nil {
		return fmt.Errorf("open SSH configuration root: %w", err)
	}
	defer func() { _ = root.Close() }()

	if err := fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk SSH configuration: %w", walkErr)
		}
		if entry.IsDir() {
			// #nosec G302 -- directories require execute permission for traversal.
			if err := root.Chmod(path, 0o700); err != nil {
				return fmt.Errorf("secure SSH directory: %w", err)
			}
			return nil
		}
		if err := root.Chmod(path, 0o600); err != nil {
			return fmt.Errorf("secure SSH file: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("secure SSH configuration tree: %w", err)
	}
	return nil
}

// CopyAndFixSSHKeys copies the authorized_keys file from the linux folder to root home and fixes all permissions.
func CopyAndFixSSHKeys(reverse bool) error {
	if unix.Access("/", unix.W_OK) != nil {
		if err := syscall.Mount("/", "/", "", syscall.MS_REMOUNT, ""); err != nil {
			return fmt.Errorf("remount root filesystem writable: %w", err)
		}

		defer func() {
			_ = syscall.Mount("/", "/", "", syscall.MS_REMOUNT|syscall.MS_RDONLY, "")
		}()
	}

	err := os.MkdirAll(config.SSHConfigFolder, 0o700)
	if err != nil {
		return fmt.Errorf("create SSH configuration directory: %w", err)
	}

	if reverse {
		err = utils.CopyFile(config.SSHKeysFile, config.UserSSHKeysFile)
	} else {
		err = utils.CopyFile(config.UserSSHKeysFile, config.SSHKeysFile)
	}
	if err != nil {
		return fmt.Errorf("copy SSH keys: %w", err)
	}

	modTime := time.Now()
	err = os.Chtimes(config.SSHKeysFile, modTime, modTime)
	if err != nil {
		return fmt.Errorf("update SSH key timestamp: %w", err)
	}
	err = os.Chtimes(config.UserSSHKeysFile, modTime, modTime)
	if err != nil {
		return fmt.Errorf("update user SSH key timestamp: %w", err)
	}

	return FixRootSSHPerms()
}
