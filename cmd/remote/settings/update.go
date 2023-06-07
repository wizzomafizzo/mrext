package settings

import (
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
