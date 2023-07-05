package mister

import (
	"context"
	"github.com/libp2p/zeroconf/v2"
	"github.com/txn2/txeh"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	DefaultHostname = "MiSTer"
	MdnsServiceName = "_mister-remote._tcp"
	mdnsPort        = 5353
	mdnsTTL         = 120
	startRetries    = 30
	discoveryTime   = 15 * time.Second
	browseInterval  = 1 * time.Minute
)

type MdnsClient struct {
	Hostname string
	Version  string
	IP       string
}

type MdnsService struct {
	mu      sync.Mutex
	Active  bool
	Clients []MdnsClient
}

func (s *MdnsService) AddClient(client MdnsClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Clients = append(s.Clients, client)
	s.Active = true
}

func (s *MdnsService) ClearClients() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Clients = []MdnsClient{}
}

func (s *MdnsService) GetClients() []MdnsClient {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Clients
}

func (s *MdnsService) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Active
}

func (s *MdnsService) SetActive(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Active = active
}

var Mdns = &MdnsService{
	Active:  false,
	Clients: []MdnsClient{},
}

func browseMdns(logger *service.Logger) {
	Mdns.ClearClients()

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

			Mdns.AddClient(MdnsClient{
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
		MdnsServiceName,
		"local.",
		entries,
		zeroconf.SelectIPTraffic(zeroconf.IPv4),
	)
	if err != nil {
		logger.Error("error during mdns browse: %s", err)
	}

	<-ctx.Done()
}

func startMdns(logger *service.Logger, appVersion string) (func() error, error) {
	if Mdns.IsActive() {
		return nil, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	logger.Info("registering mdns service with hostname: %s", hostname)
	server, err := zeroconf.Register(
		"MiSTer Remote ("+hostname+")",
		MdnsServiceName,
		"local.",
		mdnsPort,
		[]string{"version=" + appVersion},
		nil,
		zeroconf.TTL(mdnsTTL),
	)
	if err != nil {
		return nil, err
	} else {
		Mdns.SetActive(true)
	}

	logger.Info("starting network discovery service")
	browseMdns(logger)
	ticker := time.NewTicker(browseInterval)
	go func() {
		for range ticker.C {
			browseMdns(logger)
		}
	}()

	return func() error {
		ticker.Stop()
		server.Shutdown()
		Mdns.ClearClients()
		Mdns.SetActive(false)
		return nil
	}, nil
}

// TryStartMdns will attempt to start the mDNS service, retrying multiple times if it fails. This is because a script
// may be run at boot time before the network is available.
func TryStartMdns(logger *service.Logger, appVersion string) func() error {
	// TODO: allow a hook function on successful browse
	retries := 0
	for {
		stop, err := startMdns(logger, appVersion)
		if err == nil {
			return stop
		} else {
			if retries >= startRetries {
				logger.Error("failed to start mdns service, giving up: %s", err)
				return nil
			} else {
				retries++
				logger.Error("failed to start mdns service, retrying: %s", err)
				time.Sleep(time.Second)
			}
		}
	}
}

// UpdateHostname updates all hostname related files with the new hostname and refreshes it in kernel memory.
func UpdateHostname(newHostname string, writeProc bool) error {
	// TODO: also update the linux/hostname file and linux/hosts file
	procHostnameFile := "/proc/sys/kernel/hostname"
	hostnameFile := "/etc/hostname"
	localIp := "127.0.1.1"

	if newHostname == "" {
		newHostname = DefaultHostname
	}

	currentHostnameData, err := os.ReadFile(hostnameFile)
	if err != nil {
		return err
	}

	currentHostname := string(currentHostnameData)

	if currentHostname == newHostname {
		// no change required
		return nil
	}

	// remount root as read/write
	err = syscall.Mount("/", "/", "", syscall.MS_REMOUNT, "")
	if err != nil {
		return err
	}

	defer func() {
		// revert root to read-only
		_ = syscall.Mount("/", "/", "", syscall.MS_REMOUNT|syscall.MS_RDONLY, "")
	}()

	// update hostname file
	err = os.WriteFile(hostnameFile, []byte(newHostname), 0644)
	if err != nil {
		return err
	}

	// update hosts file
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		return err
	}

	hosts.RemoveHost(strings.ToLower(currentHostname))
	hosts.AddHost(localIp, strings.ToLower(newHostname))

	err = hosts.Save()
	if err != nil {
		return err
	}

	// write new hostname to proc
	if writeProc {
		err = os.WriteFile(procHostnameFile, []byte(newHostname), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

var ethAddrArg = regexp.MustCompile(`ethaddr=([0-9a-fA-F]{2}(:[0-9a-fA-F]{2}){5})`)

func readUBootConfig() (string, error) {
	uBootConfigData, err := os.ReadFile(config.UBootConfigFile)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return string(uBootConfigData), nil
}

// GetConfiguredMacAddress returns the ethernet MAC address configured in the u-boot.txt file, if available.
func GetConfiguredMacAddress() (string, error) {
	uBootConfig, err := readUBootConfig()
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(uBootConfig, "\n") {
		if ethAddrArg.MatchString(line) {
			return ethAddrArg.FindStringSubmatch(line)[1], nil
		}
	}

	return "", nil
}

// UpdateConfiguredMacAddress updates the ethernet MAC address configured in the u-boot.txt file. Setting a new one if
// it doesn't exist, or updating the existing one. Any existing u-boot.txt arguments are preserved.
func UpdateConfiguredMacAddress(newMacAddress string) error {
	uBootConfig, err := readUBootConfig()
	if err != nil {
		return err
	}

	uBootConfig = ethAddrArg.ReplaceAllString(uBootConfig, "")

	if newMacAddress != "" {
		_, err = net.ParseMAC(newMacAddress)
		if err != nil {
			return err
		}

		uBootConfig += " ethaddr=" + newMacAddress
	}

	return os.WriteFile(config.UBootConfigFile, []byte(uBootConfig), 0644)
}

func FixRootSSHPerms() error {
	err := os.Chmod(config.SSHConfigFolder, 0700)
	if err != nil {
		return err
	}

	return filepath.Walk(config.SSHConfigFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.Chmod(path, 0700)
		} else {
			return os.Chmod(path, 0600)
		}
	})
}

// CopyAndFixSSHKeys copies the authorized_keys file from the linux folder to root home and fixes all permissions.
func CopyAndFixSSHKeys(reverse bool) error {
	err := syscall.Mount("/", "/", "", syscall.MS_REMOUNT, "")
	if err != nil {
		return err
	}

	defer func() {
		_ = syscall.Mount("/", "/", "", syscall.MS_REMOUNT|syscall.MS_RDONLY, "")
	}()

	err = os.MkdirAll(config.SSHConfigFolder, 0700)
	if err != nil {
		return err
	}

	if reverse {
		err = utils.CopyFile(config.SSHKeysFile, config.UserSSHKeysFile)
	} else {
		err = utils.CopyFile(config.UserSSHKeysFile, config.SSHKeysFile)
	}
	if err != nil {
		return err
	}

	return FixRootSSHPerms()
}
