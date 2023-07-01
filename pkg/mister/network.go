package mister

import (
	"context"
	"github.com/libp2p/zeroconf/v2"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"os"
	"sync"
	"time"
)

const (
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
				version = entry.Text[0]
			}

			ip := ""
			if len(entry.AddrIPv4) > 0 {
				ip = entry.AddrIPv4[0].String()
			}

			Mdns.AddClient(MdnsClient{
				Hostname: entry.HostName,
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

func TryStartMdns(logger *service.Logger, appVersion string) func() error {
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
