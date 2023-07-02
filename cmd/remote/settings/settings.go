package settings

import (
	"encoding/json"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"net/http"
	"os/exec"
	"sync"
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
