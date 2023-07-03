package settings

import (
	"encoding/json"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

type UpdateProgress struct {
	mu      sync.Mutex
	Process *exec.Cmd
}

func (p *UpdateProgress) SetProcess(cmd *exec.Cmd) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Process = cmd
}

func (p *UpdateProgress) GetProcess() *exec.Cmd {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Process
}

var updateProgressInstance = &UpdateProgress{}

func HandleRestartRemote(cfg *config.UserConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command(cfg.AppPath, "-service", "restart")
		err := cmd.Start()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type ListPeersPayloadClient struct {
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
	IP       string `json:"ip"`
}

type ListPeersPayload struct {
	Peers []ListPeersPayloadClient `json:"peers"`
}

func HandleListPeers(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		peers := mister.Mdns.GetClients()

		payload := ListPeersPayload{
			Peers: make([]ListPeersPayloadClient, len(peers)),
		}

		for i, peer := range peers {
			payload.Peers[i] = ListPeersPayloadClient{
				Hostname: peer.Hostname,
				Version:  peer.Version,
				IP:       peer.IP,
			}
		}

		err := json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("encode list peers response: %s", err)
			return
		}
	}
}

type HandleSystemInfoPayloadDisk struct {
	Path  string `json:"path"`
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
}

type HandleSystemInfoPayload struct {
	IPs      []string `json:"ips"`
	Hostname string   `json:"hostname"`
	DNS      string   `json:"dns"`
	Version  string   `json:"version"`
	Updated  string   `json:"updated"`
}

func getNetworkIps() []string {
	ips := make([]string, 0)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, addr := range addrs {
		ip, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		if ip.IP.To4() == nil || ip.IP.IsLoopback() || ip.IP.IsMulticast() {
			continue
		}

		ips = append(ips, ip.IP.String())
	}

	return ips
}

func HandleSystemInfo(logger *service.Logger, cfg *config.UserConfig, appVer string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = ""
		}

		dns := ""
		if cfg.Remote.MdnsService {
			dns = hostname + ".local"
		}

		ips := getNetworkIps()

		updatedTime, err := mister.GetLastUpdateTime()
		updated := ""
		if err == nil {
			updated = updatedTime.Format(time.RFC3339)
		}

		payload := HandleSystemInfoPayload{
			IPs:      ips,
			Hostname: hostname,
			DNS:      dns,
			Version:  appVer,
			Updated:  updated,
		}

		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("encode system info response: %s", err)
			return
		}
	}
}

func HandleReboot(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command("reboot")
		err := cmd.Start()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("reboot: %s", err)
			return
		}
	}
}
