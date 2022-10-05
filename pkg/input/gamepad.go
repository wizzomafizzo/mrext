package input

/*
 #include <linux/input.h>
*/
import "C"

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type jsEvent struct {
	Timestamp uint32
	Value     int16
	Type      uint8
	Number    uint8
}

type evdevEvent struct {
	Time1 uint32
	Time2 uint32
	Type  uint16
	Code  uint16
	Value int32
}

const (
	typeButton uint8 = 0x01
	typeAxis   uint8 = 0x02
	typeInit   uint8 = 0x80
)

func ioctl(fd, cmd, ptr uintptr) error {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if e != 0 {
		return e
	}
	return nil
}

func readEvent(devFile string) {
	file, err := os.OpenFile(devFile, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	// ioctl(file, C.EVIOCGRAB, 0)

	buf := make([]byte, 16)
	_, err = file.Read(buf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x", buf)
}

func readGrabbedJoyState(devFile string) []jsEvent {
	fd, err := syscall.Open(devFile, syscall.O_RDONLY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 256)
	_, err = syscall.Read(fd, buf)
	if err != nil {
		panic(err)
	}

	syscall.Close(fd)

	events := make([]jsEvent, 0)
	for i := 0; i < len(buf); i = i + 8 {
		bs := buf[i : i+8]
		rbuf := bytes.NewReader(bs)

		rbuf.Seek(0, 0)
		var e jsEvent
		err = binary.Read(rbuf, binary.LittleEndian, &e)
		if err != nil {
			panic(err)
		}

		if e.Timestamp == 0 {
			continue
		} else {
			events = append(events, e)
		}
	}

	return events
}

func grabbedReadLoop() {
	pollRate := time.Millisecond * 100
	devFile := "/dev/input/js0"
	state := readGrabbedJoyState(devFile)

	for {
		newState := readGrabbedJoyState(devFile)
		if len(newState) != len(state) {
			panic("number of inputs does not match previous state")
		}

		for i := 0; i < len(newState); i++ {
			var eventType string
			if newState[i].Type&typeButton == typeButton {
				eventType = "button"
			} else {
				eventType = "axis"
			}

			if newState[i].Value != state[i].Value {
				fmt.Printf("state changed %s: %d %d\n", eventType, newState[i].Number, newState[i].Value)
			}
		}

		state = newState
		time.Sleep(pollRate)
	}
}

type Gamepad struct {
	Name    string
	EvDev   string
	JsDev   string
	SysFs   string
	Product string
	Vendor  string
	Mac     string
	Battery int
}

// Try to find battery level of gamepad, returns -1 if not found.
func getBatteryLevel(gamepad Gamepad) int {
	// TODO: worth getting the status here? charging/discharging?

	// standard bluetooth device
	ps, err := os.ReadDir(filepath.Join(gamepad.SysFs, "device", "power_supply"))
	if err == nil {
		for _, p := range ps {
			bat, err := os.ReadFile(filepath.Join(gamepad.SysFs, "device", "power_supply", p.Name(), "capacity"))
			if err != nil {
				fmt.Println(err)
				continue
			}

			level, err := strconv.Atoi(strings.TrimSpace(string(bat)))
			if err != nil {
				fmt.Println(err)
				continue
			}

			return level
		}
	} else {
		fmt.Println(err)
	}

	// xbox
	btctl := exec.Command("bluetoothctl", "info", gamepad.Mac)
	btctlOut, err := btctl.Output()
	if err != nil {
		fmt.Println(err)
		return -1
	} else {
		re := regexp.MustCompile(`Battery Percentage: .+ \((\d+)\)`)
		matches := re.FindStringSubmatch(string(btctlOut))
		if len(matches) == 2 {
			level, err := strconv.Atoi(matches[1])
			if err != nil {
				fmt.Println(err)
				return -1
			}

			return level
		}
	}

	return -1
}

// Return a list of all connected gamepads.
func GetGamepads() ([]Gamepad, error) {
	devices, err := os.Open("/proc/bus/input/devices")
	if err != nil {
		return nil, err
	}
	defer devices.Close()

	var gamepads []Gamepad
	var gamepad Gamepad

	sc := bufio.NewScanner(devices)
	for sc.Scan() {
		line := sc.Text()

		if line == "" {
			// FIXME: not sure how reliable this is to say what a gamepad is
			if gamepad.JsDev != "" {
				gamepads = append(gamepads, gamepad)
			}
			gamepad = Gamepad{}
			continue
		}

		if line[0] == 'I' {
			items := strings.Split(line[3:], " ")
			for _, item := range items {
				def := strings.Split(item, "=")
				if def[0] == "Vendor" {
					gamepad.Vendor = strings.ToLower(def[1])
				} else if def[0] == "Product" {
					gamepad.Product = strings.ToLower(def[1])
				}
			}
		}

		if line[0] == 'N' {
			// FIXME: this always assumes the name is quoted
			gamepad.Name = line[9 : len(line)-1]
		}

		if line[0] == 'S' {
			gamepad.SysFs = "/sys" + line[9:]
		}

		if line[0] == 'U' {
			gamepad.Mac = strings.ToLower(line[8:])
		}

		if line[0] == 'H' {
			devs := strings.Split(line[12:], " ")
			for _, dev := range devs {
				if strings.HasPrefix(dev, "event") {
					gamepad.EvDev = "/dev/input/" + dev
				} else if strings.HasPrefix(dev, "js") {
					gamepad.JsDev = "/dev/input/" + dev
				}
			}
		}
	}

	for i, gp := range gamepads {
		gamepads[i].Battery = getBatteryLevel(gp)
	}

	return gamepads, nil
}
