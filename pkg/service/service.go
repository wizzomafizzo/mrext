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
	return readServicePID(s.pidFilePath())
}

func readServicePID(path string) (int, error) {
	// #nosec G304 -- configured PID path in production; temporary fixture in tests.
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read service PID: %w", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse service PID: %w", err)
	}
	if pid <= 1 {
		return 0, errors.New("invalid service PID: must be greater than 1")
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

	defer func() { _ = process.Release() }()
	return s.matchesDaemon(pid) && process.Signal(syscall.Signal(0)) == nil
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

	s.stop = stop
	s.setupStopService()

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

	// Write the copy beside its final name and rename it into place. Opening
	// the destination with O_TRUNC fails with ETXTBSY when a daemon is still
	// running from it, which happens whenever the PID file went missing while
	// the process lived, and reported "text file busy" rather than anything
	// about the real cause. pkg/bgm/daemon.go already does it this way.
	tempPath := filepath.Join(config.TempFolder, filepath.Base(binPath))
	// #nosec G304,G703 -- staged beside the controlled temp path above.
	tempFile, err := os.CreateTemp(config.TempFolder, "."+filepath.Base(binPath)+"-*")
	if err != nil {
		return fmt.Errorf("error creating temp binary: %w", err)
	}
	stagedPath := tempFile.Name()
	published := false
	defer func() {
		_ = tempFile.Close()
		if !published {
			_ = os.Remove(stagedPath)
		}
	}()

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
	// #nosec G302 -- the copied service binary must be executable.
	if chmodErr := os.Chmod(stagedPath, 0o755); chmodErr != nil {
		return fmt.Errorf("make temporary binary executable: %w", chmodErr)
	}
	// #nosec G703 -- destination is the controlled temp path built above.
	if renameErr := os.Rename(stagedPath, tempPath); renameErr != nil {
		return fmt.Errorf("publish temporary binary: %w", renameErr)
	}
	published = true

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
	pid, err := s.Pid()
	if err != nil {
		return fmt.Errorf("read service PID: %w", err)
	}

	// No PID file at all is the ordinary stopped case, not a mismatched PID.
	if pid == 0 {
		return fmt.Errorf("%s service not running", s.Name)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find service process: %w", err)
	}

	defer func() { _ = process.Release() }()
	if !s.matchesDaemon(pid) {
		return fmt.Errorf("%s PID does not identify its service daemon", s.Name)
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("signal service process: %w", err)
	}

	return nil
}

// stopTimeout bounds how long Restart waits for the previous daemon to exit.
// A variable so tests need not wait it out.
var stopTimeout = 20 * time.Second

// Restart stops the running service and starts it again.
//
// The wait used to be an unbounded "for s.Running() { sleep }", so a daemon
// wedged in its own shutdown, such as a blocked listener or file watcher
// close, meant restart never returned and never timed out. Give up waiting
// politely after stopTimeout and escalate to SIGKILL.
func (s *Service) Restart() error {
	if s.Running() {
		if err := s.Stop(); err != nil {
			return err
		}
	}

	deadline := time.Now().Add(stopTimeout)
	for s.Running() {
		if time.Now().After(deadline) {
			if err := s.kill(); err != nil {
				return err
			}
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	return s.Start()
}

// kill sends SIGKILL to a daemon that did not honour SIGTERM, then clears the
// PID file so the restart is not blocked by the corpse.
func (s *Service) kill() error {
	pid, err := s.Pid()
	if err != nil {
		return fmt.Errorf("read service PID: %w", err)
	}
	if pid == 0 {
		return nil
	}
	if !s.matchesDaemon(pid) {
		return nil
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find service process: %w", err)
	}
	defer func() { _ = process.Release() }()
	s.Logger.Warn("%s did not stop in %s, killing it", s.Name, stopTimeout)
	if err := process.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("kill service process: %w", err)
	}
	return nil
}

// exitOnError reports a failed service command on the console as well as in
// the log, then exits. Logging alone left the user with a silent exit 1 and no
// way to know they had to read /tmp/<app>.log to find out why.
func (s *Service) exitOnError(err error) {
	if err != nil {
		s.Logger.Error("%s", err)
		_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", s.Name, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func (s *Service) ServiceHandler(cmd *string) {
	switch *cmd {
	case "exec":
		s.startService()
		os.Exit(0)
	case "start":
		s.exitOnError(s.Start())
	case "stop":
		s.exitOnError(s.Stop())
	case "restart":
		s.exitOnError(s.Restart())
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
