package service

import (
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
	Name   string
	Logger *Logger
	daemon bool
	start  ServiceEntry
	stop   func() error
}

type ServiceArgs struct {
	Name     string
	Logger   *Logger
	Entry    ServiceEntry
	NoDaemon bool
}

func NewService(args ServiceArgs) (*Service, error) {
	if args.Name == "" {
		return nil, fmt.Errorf("service name is required")
	}

	if args.Logger == nil {
		return nil, fmt.Errorf("service logger is required")
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
	err := os.WriteFile(s.pidFilePath(), []byte(fmt.Sprintf("%d", pid)), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) removePidFile() error {
	err := os.Remove(s.pidFilePath())
	if err != nil {
		return err
	}
	return nil
}

// Return the process ID of the current running service daemon.
func (s *Service) Pid() (int, error) {
	pidPath := fmt.Sprintf(config.PidFileTemplate, s.Name)
	pid := 0

	if _, err := os.Stat(pidPath); err == nil {
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

	if s.daemon {
		<-make(chan struct{})
	} else {
		err := s.stopService()
		if err != nil {
			os.Exit(1)
		}

		os.Exit(0)
	}
}

// Start a new service daemon in the background.
func (s *Service) Start() error {
	if s.Running() {
		return fmt.Errorf("%s service already running", s.Name)
	}

	// create a copy in binary in tmp so the original can be updated
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting absolute binary path: %w", err)
	}

	binFile, err := os.Open(binPath)
	if err != nil {
		return fmt.Errorf("error opening binary: %w", err)
	}

	tempPath := filepath.Join(config.TempFolder, filepath.Base(binPath))
	tempFile, err := os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("error creating temp binary: %w", err)
	}

	_, err = io.Copy(tempFile, binFile)
	if err != nil {
		return fmt.Errorf("error copying binary to temp: %w", err)
	}

	tempFile.Close()
	binFile.Close()

	cmd := exec.Command(tempPath, "-service", "exec", "&")

	// point new binary to existing config file
	configPath := filepath.Join(filepath.Dir(binPath), s.Name+".ini")
	appPath, _ := os.Executable()
	if _, err := os.Stat(configPath); err == nil {
		env := os.Environ()
		cmd.Env = env
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", config.UserConfigEnv, configPath))
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", config.UserAppPathEnv, appPath))
	}

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
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Restart() error {
	err := s.Stop()
	if err != nil {
		return err
	}

	for s.Running() {
		time.Sleep(1 * time.Second)
	}

	err = s.Start()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ServiceHandler(cmd *string) {
	if *cmd == "exec" {
		s.startService()
		os.Exit(0)
	} else if *cmd == "start" {
		err := s.Start()
		if err != nil {
			s.Logger.Error(err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	} else if *cmd == "stop" {
		err := s.Stop()
		if err != nil {
			s.Logger.Error(err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	} else if *cmd == "restart" {
		err := s.Restart()
		if err != nil {
			s.Logger.Error(err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	} else if *cmd == "status" {
		if s.Running() {
			fmt.Printf("%s service running\n", s.Name)
		} else {
			fmt.Printf("%s service not running\n", s.Name)
		}

		os.Exit(0)
	} else if *cmd != "" {
		fmt.Printf("Invalid service command: %s", *cmd)
		os.Exit(1)
	}
}
