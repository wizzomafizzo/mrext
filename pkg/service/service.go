//go:build linux

package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

type ServiceEntry func() (func() error, error)

type Service struct {
	Logger *Logger
	start  ServiceEntry
	stop   func() error
	Name   string
	daemon bool
}

type ServiceArgs struct {
	Logger   *Logger
	Entry    ServiceEntry
	Name     string
	NoDaemon bool
}

func NewService(args ServiceArgs) (*Service, error) {
	if args.Name == "" {
		return nil, errors.New("service name is required")
	}

	if args.Logger == nil {
		return nil, errors.New("service logger is required")
	}

	return &Service{
		Name:   args.Name,
		Logger: args.Logger,
		daemon: !args.NoDaemon,
		start:  args.Entry,
	}, nil
}

func (s *Service) pidFilePath() string {
	return fmt.Sprintf(config.PidFileTemplate, s.Name)
}

// Create new PID file using current process PID.
func (s *Service) createPidFile() error {
	pid := os.Getpid()
	// #nosec G306 -- service PID files must remain world-readable.
	err := os.WriteFile(s.pidFilePath(), []byte(strconv.Itoa(pid)), 0o644)
	if err != nil {
		return fmt.Errorf("write service PID file: %w", err)
	}
	return nil
}

func (s *Service) removePidFile() error {
	err := os.Remove(s.pidFilePath())
	if err != nil {
		return fmt.Errorf("remove service PID file: %w", err)
	}
	return nil
}

// Return the process ID of the current running service daemon.
func (s *Service) Pid() (int, error) {
	pidPath := fmt.Sprintf(config.PidFileTemplate, s.Name)
	pid := 0

	if _, err := os.Stat(pidPath); err == nil {
		// #nosec G304 -- path is derived from configured service name and PID template.
		pidFile, err := os.ReadFile(pidPath)
		if err != nil {
			return pid, fmt.Errorf("error reading pid file: %w", err)
		}

		pidInt, err := strconv.Atoi(string(pidFile))
		if err != nil {
			return pid, fmt.Errorf("error parsing pid: %w", err)
		}

		pid = pidInt
	}

	return pid, nil
}

// Returns true if the service is running.
func (s *Service) Running() bool {
	pid, err := s.Pid()
	if err != nil {
		return false
	}

	if pid == 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))

	return err == nil
}

func (s *Service) stopService() error {
	s.Logger.Info("stopping %s service", s.Name)

	err := s.stop()
	if err != nil {
		s.Logger.Error("error stopping %s service: %s", s.Name, err)
		return err
	}

	err = s.removePidFile()
	if err != nil {
		s.Logger.Error("error removing pid file: %s", err)
		return err
	}

	// remove temporary binary
	tempPath, err := os.Executable()
	if err != nil {
		s.Logger.Error("error getting executable path: %s", err)
	} else if strings.HasPrefix(tempPath, config.TempFolder) {
		err = os.Remove(tempPath)
		if err != nil {
			s.Logger.Error("error removing temporary binary: %s", err)
		}
	}

	return nil
}

// Set up signal handler to stop service on SIGINT or SIGTERM. Exits the application on signal.
func (s *Service) setupStopService() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		err := s.stopService()
		if err != nil {
			os.Exit(1)
		}

		os.Exit(0)
	}()
}

// Starts the service and blocks until the service is stopped.
func (s *Service) startService() {
	if s.Running() {
		s.Logger.Error("%s service already running", s.Name)
		os.Exit(1)
	}

	s.Logger.Info("starting %s service", s.Name)

	err := s.createPidFile()
	if err != nil {
		s.Logger.Error("error creating pid file: %s", err)
		os.Exit(1)
	}

	err = SetNice()
	if err != nil {
		s.Logger.Error("error setting nice level: %s", err)
	}

	stop, err := s.start()
	if err != nil {
		s.Logger.Error("error starting service: %s", err)

		err = s.removePidFile()
		if err != nil {
			s.Logger.Error("error removing pid file: %s", err)
		}

		os.Exit(1)
	}

	s.setupStopService()
	s.stop = stop

	if !s.daemon {
		err := s.stopService()
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	<-make(chan struct{})
}

// Start a new service daemon in the background.
func (s *Service) Start() error {
	if s.Running() {
		return fmt.Errorf("%s service already running", s.Name)
	}

	// create a copy in binary in tmp so the original can be updated
	binPath := ""
	appPath := os.Getenv(config.UserAppPathEnv)
	if appPath != "" {
		binPath = appPath
	} else {
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("error getting absolute binary path: %w", err)
		}
		binPath = exePath
	}

	// #nosec G304,G703 -- binary path is current executable or explicit application path.
	binFile, err := os.Open(binPath)
	if err != nil {
		return fmt.Errorf("error opening binary: %w", err)
	}
	defer func() { _ = binFile.Close() }()

	tempPath := filepath.Join(config.TempFolder, filepath.Base(binPath))
	// #nosec G302,G304,G703 -- copied service binary must be executable at its controlled temp path.
	tempFile, err := os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("error creating temp binary: %w", err)
	}
	defer func() { _ = tempFile.Close() }()

	_, err = io.Copy(tempFile, binFile)
	if err != nil {
		return fmt.Errorf("error copying binary to temp: %w", err)
	}

	if closeErr := tempFile.Close(); closeErr != nil {
		return fmt.Errorf("close temporary binary: %w", closeErr)
	}
	if closeErr := binFile.Close(); closeErr != nil {
		return fmt.Errorf("close source binary: %w", closeErr)
	}

	// #nosec G204,G702 -- executable is controlled service copy created above.
	cmd := exec.CommandContext(context.Background(), tempPath, "-service", "exec", "&")
	env := os.Environ()
	cmd.Env = env

	// point new binary to existing config file
	configPath := filepath.Join(filepath.Dir(binPath), s.Name+".ini")

	// #nosec G703 -- path is derived from explicit application path and service name.
	if _, statErr := os.Stat(configPath); statErr == nil {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", config.UserConfigEnv, configPath))
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", config.UserAppPathEnv, binPath))

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting %s service: %w", s.Name, err)
	}

	return nil
}

// Stop the service daemon.
func (s *Service) Stop() error {
	if !s.Running() {
		return fmt.Errorf("%s service not running", s.Name)
	}

	pid, err := s.Pid()
	if err != nil {
		return fmt.Errorf("read service PID: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find service process: %w", err)
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("signal service process: %w", err)
	}

	return nil
}

func (s *Service) Restart() error {
	if s.Running() {
		err := s.Stop()
		if err != nil {
			return err
		}
	}

	for s.Running() {
		time.Sleep(1 * time.Second)
	}

	err := s.Start()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ServiceHandler(cmd *string) {
	switch *cmd {
	case "exec":
		s.startService()
		os.Exit(0)
	case "start":
		if err := s.Start(); err != nil {
			s.Logger.Error("%s", err)
			os.Exit(1)
		}
		os.Exit(0)
	case "stop":
		if err := s.Stop(); err != nil {
			s.Logger.Error("%s", err)
			os.Exit(1)
		}
		os.Exit(0)
	case "restart":
		if err := s.Restart(); err != nil {
			s.Logger.Error("%s", err)
			os.Exit(1)
		}
		os.Exit(0)
	case "status":
		if s.Running() {
			_, _ = fmt.Printf("%s service running\n", s.Name)
		} else {
			_, _ = fmt.Printf("%s service not running\n", s.Name)
		}
		os.Exit(0)
	case "":
		return
	default:
		_, _ = fmt.Printf("Invalid service command: %s", *cmd)
		os.Exit(1)
	}
}
