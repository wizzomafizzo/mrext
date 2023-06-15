package settings

import (
	"github.com/wizzomafizzo/mrext/pkg/config"
	"net/http"
	"os/exec"
	"path/filepath"
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

func HandleRestartRemote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: this will break if the remote.sh script is launched from somewhere else
		//       it probably will rarely ever happen, but the path should be found dynamically
		//       it can't be found from memory because service is launched from tmp
		path := filepath.Join(config.ScriptsFolder, "remote.sh")
		cmd := exec.Command(path, "-service", "restart")
		err := cmd.Start()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
