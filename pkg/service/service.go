package service

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

type Service struct {
	Name   string
	Logger *Logger
	start  func() error
	stop   func() error
}

type ServiceArgs struct {
	Name   string
	Logger *Logger
	Start  func() error
	Stop   func() error
}

func NewService(args ServiceArgs) (*Service, error) {
	if args.Name == "" {
		return nil, fmt.Errorf("service name is required")
	}

	if args.Logger == nil {
		return nil, fmt.Errorf("service logger is required")
	}

	return &Service{
		Name:  args.Name,
		start: args.Start,
		stop:  args.Stop,
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

// Set up signal handler to stop service on SIGINT or SIGTERM. Exits the application on signal.
func (s *Service) setupStopService() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		err := s.stop()
		if err != nil {
			s.Logger.Error("error stopping %s service: %s", s.Name, err)
			os.Exit(1)
		}

		err = s.removePidFile()
		if err != nil {
			s.Logger.Error("error removing pid file: %s", err)
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

	err := s.createPidFile()
	if err != nil {
		s.Logger.Error("error creating pid file: %s", err)
		os.Exit(1)
	}

	s.setupStopService()

	err = s.start()
	if err != nil {
		s.Logger.Error("error starting service: %s", err)

		err = s.removePidFile()
		if err != nil {
			s.Logger.Error("error removing pid file: %s", err)
		}

		os.Exit(1)
	}

	<-make(chan struct{})
}

// Start a new service daemon in the background.
func (s *Service) Start() error {
	if s.Running() {
		return fmt.Errorf("%s service already running", s.Name)
	}

	err := exec.Command(os.Args[0], "-service", "exec", "&").Start()
	if err != nil {
		return fmt.Errorf("error starting % service: %w", s.Name, err)
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

func (s *Service) FlagHandler(cmd *string) {
	if *cmd == "exec" {
		s.startService()
	} else if *cmd == "start" {
		err := s.Start()
		if err != nil {
			s.Logger.Error(err.Error())
			os.Exit(1)
		}
	} else if *cmd == "stop" {
		err := s.Stop()
		if err != nil {
			s.Logger.Error(err.Error())
			os.Exit(1)
		}
	} else if *cmd == "restart" {
		err := s.Stop()
		if err != nil {
			s.Logger.Error(err.Error())
			os.Exit(1)
		}

		for s.Running() {
			time.Sleep(1 * time.Second)
		}

		err = s.Start()
		if err != nil {
			s.Logger.Error(err.Error())
			os.Exit(1)
		}
	} else if *cmd == "status" {
		if s.Running() {
			fmt.Printf("%s service running\n", s.Name)
		} else {
			fmt.Printf("%s service not running\n", s.Name)
		}
	}

	os.Exit(0)
}
