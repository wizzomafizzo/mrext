package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	gc "github.com/rthornton128/goncurses"
	"github.com/wizzomafizzo/mrext/pkg/curses"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"

	"github.com/clausecker/nfc/v2"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

// TODO: something like the nfc-list utility so new users with unsupported readers can help identify them
// TODO: play sound using go library
// TODO: a -test command to see what the result of an NDEF would be
// TODO: would it be possible to unlock the OSD with a card?
// TODO: more concrete amiibo support
// TODO: create a test web nfc reader in separate github repo, hosted on pages
// TODO: use a tag to signal that that next tag should have the active game written to it
// TODO: option to use search.db instead of on demand index for random
// TODO: check if a game is launched while the gui is open and close it (may not be possible)
// TODO: clean up mgl files in tmp
// TODO: multi line text support on log view

const (
	appName            = "nfc"
	connectMaxTries    = 10
	timesToPoll        = 20
	periodBetweenPolls = 300 * time.Millisecond
	periodBetweenLoop  = 300 * time.Millisecond
	timeToForgetCard   = 5 * time.Second
	successPath        = config.TempFolder + "/success.wav"
	failPath           = config.TempFolder + "/fail.wav"
)

var (
	supportedCardTypes = []nfc.Modulation{
		{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106},
	}
	logger = service.NewLogger(appName)
)

type Card struct {
	CardType string
	UID      string
	Text     string
	ScanTime time.Time
}

type ServiceState struct {
	mu              sync.Mutex
	activeCard      Card
	lastScanned     Card
	stopService     bool
	disableLauncher bool
	dbLoadTime      time.Time
	uidMap          map[string]string
	textMap         map[string]string
}

func (s *ServiceState) SetActiveCard(card Card) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeCard = card
	if s.activeCard.UID != "" {
		s.lastScanned = card
	}
}

func (s *ServiceState) GetActiveCard() Card {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeCard
}

func (s *ServiceState) GetLastScanned() Card {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastScanned
}

func (s *ServiceState) StopService() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopService = true
}

func (s *ServiceState) ShouldStopService() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopService
}

func (s *ServiceState) DisableLauncher() {
	// TODO: implement this
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disableLauncher = true
}

func (s *ServiceState) EnableLauncher() {
	// TODO: implement this
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disableLauncher = false
}

func (s *ServiceState) IsLauncherDisabled() bool {
	// TODO: implement this
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disableLauncher
}

func (s *ServiceState) GetDB() (map[string]string, map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.uidMap, s.textMap
}

func (s *ServiceState) GetDBLoadTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dbLoadTime
}

func (s *ServiceState) SetDB(uidMap map[string]string, textMap map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dbLoadTime = time.Now()
	s.uidMap = uidMap
	s.textMap = textMap
}

func pollDevice(
	pnd *nfc.Device,
	activeCard Card,
) (Card, error) {
	count, target, err := pnd.InitiatorPollTarget(supportedCardTypes, timesToPoll, periodBetweenPolls)
	if err != nil && !errors.Is(err, nfc.Error(nfc.ETIMEOUT)) {
		return activeCard, err
	}

	if count <= 0 {
		if activeCard.UID != "" && time.Since(activeCard.ScanTime) > timeToForgetCard {
			logger.Info("card removed")
			activeCard = Card{}
		}

		return activeCard, nil
	}

	cardUid := getCardUID(target)
	if cardUid == "" {
		logger.Warn("unable to detect card UID: %s", target.String())
	}

	if cardUid == activeCard.UID {
		return activeCard, nil
	}

	logger.Info("card UID: %s", cardUid)

	cardType, err := getCardType(*pnd)
	if err != nil {
		logger.Error("error getting card type: %s", err)
	} else if cardType == "" {
		logger.Warn("unknown card type")
	} else {
		logger.Info("card type: %s", cardType)
	}

	blockCount := getDataAreaSize(cardType)
	record, err := readRecord(*pnd, blockCount)
	if err != nil {
		return activeCard, fmt.Errorf("error reading record: %s", err)
	}
	logger.Debug("record bytes: %s", hex.EncodeToString(record))

	tagText := parseRecordText(record)
	if tagText == "" {
		logger.Warn("no text NDEF found")
	} else {
		logger.Info("decoded text NDEF: %s", tagText)
	}

	card := Card{
		CardType: cardType,
		UID:      cardUid,
		Text:     tagText,
		ScanTime: time.Now(),
	}

	return card, nil
}

func startService(cfg *config.UserConfig) (func() error, error) {
	state := &ServiceState{}

	err := loadDatabase(state)
	if err != nil {
		logger.Error("error loading database: %s", err)
	}

	// TODO: don't want to depend on external aplay command, but i'm out of
	//       time to keep messing with this. oto/beep would not work for me
	//       and are annoying to compile statically
	sf, err := os.Create(successPath)
	if err != nil {
		logger.Error("error creating success sound file: %s", err)
	}
	_, err = sf.Write(successSound)
	if err != nil {
		logger.Error("error writing success sound file: %s", err)
	}
	_ = sf.Close()
	playSuccess := func() {
		err := exec.Command("aplay", successPath).Start()
		if err != nil {
			logger.Error("error playing success sound: %s", err)
		}
	}

	ff, err := os.Create(failPath)
	if err != nil {
		logger.Error("error creating fail sound file: %s", err)
	}
	_, err = ff.Write(failSound)
	if err != nil {
		logger.Error("error writing fail sound file: %s", err)
	}
	_ = ff.Close()
	playFail := func() {
		err := exec.Command("aplay", failPath).Start()
		if err != nil {
			logger.Error("error playing fail sound: %s", err)
		}
	}

	var closeDbWatcher func() error
	dbWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("error creating watcher: %s", err)
	} else {
		closeDbWatcher = dbWatcher.Close
	}

	go func() {
		// this turned out to be not trivial to say the least, mostly due to
		// the fact the fsnotify library does not implement the IN_CLOSE_WRITE
		// inotify event, which signals the file has finished being written
		// see: https://github.com/fsnotify/fsnotify/issues/372
		//
		// during a standard write operation, a file may emit multiple write
		// events, including when the file could be half-written
		//
		// it's also the case that editors may delete the file and create a new
		// one, which kills the active watcher
		//
		// this solution is very ugly, but it appears to work well :)
		// i think it will be sufficient for the use case, and i really like
		// this idea a lot. it's certainly preferable to the screen flicker
		// with the previous setup
		//
		// there doesn't appear to be any actively maintained wrapper for
		// inotify, so i think it would be best to write one for mrext later
		const delay = 1 * time.Second
		for {
			select {
			case event, ok := <-dbWatcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					// usually receives multiple write events, just act on the first
					if time.Since(state.GetDBLoadTime()) < delay {
						continue
					}
					time.Sleep(delay)
					logger.Info("database changed, reloading")
					err := loadDatabase(state)
					if err != nil {
						logger.Error("error loading database: %s", err)
					}
				} else if event.Has(fsnotify.Remove) {
					// editors may also delete the file on write
					time.Sleep(delay)
					_, err := os.Stat(config.NfcDatabaseFile)
					if err == nil {
						err = dbWatcher.Add(config.NfcDatabaseFile)
						if err != nil {
							logger.Error("error watching database: %s", err)
						}
						logger.Info("database changed, reloading")
						err := loadDatabase(state)
						if err != nil {
							logger.Error("error loading database: %s", err)
						}
					}
				}
			case err, ok := <-dbWatcher.Errors:
				if !ok {
					return
				}
				logger.Error("watcher error: %s", err)
			}
		}
	}()

	err = dbWatcher.Add(config.NfcDatabaseFile)
	if err != nil {
		logger.Error("error watching database: %s", err)
	}

	go func() {
		var pnd nfc.Device
		var err error

	reconnect:
		tries := 0
		for {
			// TODO: don't show every failed attempt error, tidy the log
			pnd, err = nfc.Open(cfg.NfcConfig.ConnectionString)
			if err != nil {
				logger.Error("could not open device: %s", err)
				if tries >= connectMaxTries {
					logger.Error("giving up, exiting")
					return
				}
			} else {
				break
			}
			tries++
		}

		defer func(pnd nfc.Device) {
			err := pnd.Close()
			if err != nil {
				logger.Warn("error closing device: %s", err)
			}
		}(pnd)

		if err := pnd.InitiatorInit(); err != nil {
			logger.Error("could not init initiator: %s", err)
			return
		}

		logger.Info("opened connection: %s %s", pnd, pnd.Connection())
		logger.Info("polling for %d times with %s delay", timesToPoll, periodBetweenPolls)
		var lastError time.Time

		for {
			if state.ShouldStopService() {
				break
			}

			activeCard := state.GetActiveCard()
			newScanned, err := pollDevice(&pnd, activeCard)
			if errors.Is(err, nfc.Error(nfc.EIO)) {
				logger.Error("error during poll: %s", err)
				logger.Error("fatal IO error, device was unplugged, exiting...")
				if time.Since(lastError) > 1*time.Second {
					playFail()
				}
				goto reconnect
			} else if err != nil {
				logger.Error("error during poll: %s", err)
				if time.Since(lastError) > 1*time.Second {
					playFail()
				}
				lastError = time.Now()
				goto end
			}

			state.SetActiveCard(newScanned)

			if newScanned.UID == "" || activeCard.UID == newScanned.UID {
				goto end
			}

			playSuccess()

			err = writeScanResult(newScanned)
			if err != nil {
				logger.Warn("error writing tmp scan result: %s", err)
			}

			err = launchCard(cfg, state)
			if err != nil {
				logger.Error("error launching card: %s", err)
				if time.Since(lastError) > 1*time.Second {
					playFail()
				}
				lastError = time.Now()
				goto end
			}

		end:
			time.Sleep(periodBetweenLoop)
		}
	}()

	socket, err := net.Listen("unix", config.TempFolder+"/nfc.sock")
	if err != nil {
		logger.Error("error creating socket: %s", err)
		return nil, err
	}

	go func() {
		for {
			if state.ShouldStopService() {
				break
			}

			conn, err := socket.Accept()
			if err != nil {
				logger.Error("error accepting connection: %s", err)
				return
			}

			go func(conn net.Conn) {
				logger.Debug("new socket connection")

				defer func(conn net.Conn) {
					err := conn.Close()
					if err != nil {
						logger.Warn("error closing connection: %s", err)
					}
				}(conn)

				buf := make([]byte, 4096)

				n, err := conn.Read(buf)
				if err != nil {
					logger.Error("error reading from connection: %s", err)
					return
				}

				if n == 0 {
					return
				}
				logger.Debug("received %d bytes", n)

				payload := ""

				switch strings.TrimSpace(string(buf[:n])) {
				case "status":
					lastScanned := state.GetLastScanned()
					if lastScanned.UID != "" {
						payload = fmt.Sprintf(
							"%d,%s,%s",
							lastScanned.ScanTime.Unix(),
							lastScanned.UID,
							lastScanned.Text,
						)
					} else {
						payload = "0,,"
					}
				default:
					logger.Warn("unknown command: %s", strings.TrimRight(string(buf[:n]), "\n"))
				}

				if payload == "" {
					return
				}

				_, err = conn.Write([]byte(payload))
				if err != nil {
					logger.Error("error writing to connection: %s", err)
					return
				}
			}(conn)
		}
	}()

	return func() error {
		err := socket.Close()
		if err != nil {
			logger.Warn("error closing socket: %s", err)
		}
		state.StopService()
		if closeDbWatcher != nil {
			return closeDbWatcher()
		}
		return nil
	}, nil
}

func writeScanResult(card Card) error {
	f, err := os.Create(config.NfcLastScanFile)
	if err != nil {
		return fmt.Errorf("unable to create scan result file %s: %s", config.NfcLastScanFile, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.WriteString(fmt.Sprintf("%s,%s", card.UID, card.Text))
	if err != nil {
		return fmt.Errorf("unable to write scan result file %s: %s", config.NfcLastScanFile, err)
	}

	return nil
}

func addToStartup() error {
	var startup mister.Startup

	err := startup.Load()
	if err != nil {
		return err
	}

	if !startup.Exists("mrext/" + appName) {
		err = startup.AddService("mrext/" + appName)
		if err != nil {
			return err
		}

		err = startup.Save()
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	svcOpt := flag.String("service", "", "manage nfc service (start, stop, restart, status)")
	flag.Parse()

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{})
	if err != nil {
		logger.Error("error loading user config: %s", err)
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	svc, err := service.NewService(service.ServiceArgs{
		Name:   appName,
		Logger: logger,
		Entry: func() (func() error, error) {
			return startService(cfg)
		},
	})
	if err != nil {
		logger.Error("error creating service: %s", err)
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}

	svc.ServiceHandler(svcOpt)

	interactive := true
	stdscr, err := curses.Setup()
	if err != nil {
		logger.Error("starting curses: %s", err)
		interactive = false
	}
	defer gc.End()

	if !interactive {
		err = addToStartup()
		if err != nil {
			logger.Error("error adding startup: %s", err)
			fmt.Println("Error adding to startup:", err)
		}
	} else {
		err = tryAddStartup(stdscr)
		if err != nil {
			logger.Error("error adding startup: %s", err)
		}
	}

	if !svc.Running() {
		err := svc.Start()
		if err != nil {
			logger.Error("error starting service: %s", err)
			if !interactive {
				fmt.Println("Error starting service:", err)
			}
			os.Exit(1)
		} else if !interactive {
			fmt.Println("Service started successfully.")
			os.Exit(0)
		}
	} else if !interactive {
		fmt.Println("Service is running.")
		os.Exit(0)
	}

	err = displayServiceInfo(stdscr, svc)
	if err != nil {
		logger.Error("error displaying service info: %s", err)
	}
}
