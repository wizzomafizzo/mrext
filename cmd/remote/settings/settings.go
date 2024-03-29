package settings

import (
	"embed"
	"encoding/json"
	"fmt"
	gm "github.com/c-seeger/mac-gen-go"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

func HandleRestartRemote(logger *service.Logger, cfg *config.UserConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("restart remote request")
		cmd := exec.Command(cfg.AppPath, "-service", "restart")
		err := cmd.Start()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error restarting: %s", err)
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
	Path        string `json:"path"`
	Total       uint64 `json:"total"`
	Used        uint64 `json:"used"`
	Free        uint64 `json:"free"`
	DisplayName string `json:"displayName"`
}

type HandleSystemInfoPayload struct {
	IPs      []string                      `json:"ips"`
	Hostname string                        `json:"hostname"`
	DNS      string                        `json:"dns"`
	Version  string                        `json:"version"`
	Updated  string                        `json:"updated"`
	Disks    []HandleSystemInfoPayloadDisk `json:"disks"`
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

		if ip.IP.To4() == nil {
			continue
		}

		if ip.IP.IsLoopback() || ip.IP.IsMulticast() || ip.IP.IsLinkLocalUnicast() || ip.IP.IsLinkLocalMulticast() {
			continue
		}

		ips = append(ips, ip.IP.String())
	}

	return ips
}

func getDiskInfo(cfg *config.UserConfig) ([]HandleSystemInfoPayloadDisk, error) {
	diskInfo := make([]HandleSystemInfoPayloadDisk, 0)

	mounts, err := mister.GetMounts(cfg)
	if err != nil {
		return diskInfo, err
	}

	for _, mount := range mounts {
		info, err := mister.GetDiskUsage(mount)
		if err != nil {
			return diskInfo, err
		}

		displayName := ""

		if mount == config.SdFolder {
			displayName = "SD card"
		} else if mount == config.CifsFolder {
			displayName = "Network share"
		} else {
			displayName = filepath.Base(mount)
		}

		diskInfo = append(diskInfo, HandleSystemInfoPayloadDisk{
			Path:        mount,
			Total:       info.Total,
			Used:        info.Used,
			Free:        info.Free,
			DisplayName: displayName,
		})
	}

	return diskInfo, nil
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

		diskInfo, err := getDiskInfo(cfg)
		if err != nil {
			logger.Error("error getting disk info: %s", err)
		}

		payload := HandleSystemInfoPayload{
			IPs:      ips,
			Hostname: hostname,
			DNS:      dns,
			Version:  appVer,
			Updated:  updated,
			Disks:    diskInfo,
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

type GenerateMacPayload struct {
	Mac string `json:"mac"`
}

func HandleGenerateMac(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := GenerateMacPayload{}

		ip, err := utils.GetLocalIp()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("get local ip: %s", err)
			return
		}

		prefix := gm.GenerateRandomLocalMacPrefix(true)

		suffix, err := gm.CalculateNICSufix(ip)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("generate mac: %s", err)
			return
		}

		payload.Mac = fmt.Sprintf("%s:%s", prefix, suffix)

		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("encode generate mac response: %s", err)
			return
		}
	}
}

func HandleLogoFile(logger *service.Logger, client embed.FS, cfg *config.UserConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var path string
		var data []byte
		var err error

		if cfg.Remote.CustomLogo != "" {
			path = cfg.Remote.CustomLogo
			data, err = os.ReadFile(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.Error("read custom logo file: %s", err)
				return
			}
		}

		if len(data) == 0 {
			path = "_client/build/misterlogo.svg"
			data, err = client.ReadFile(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.Error("read logo file: %s", err)
				return
			}
		}

		contentType := mime.TypeByExtension(filepath.Ext(path))
		w.Header().Set("Content-Type", contentType)

		_, err = w.Write(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("server logo file: %s", err)
			return
		}
	}
}
