package gamepad

/*
 #include <linux/input.h>
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
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

func readBuggyJoyState(devFile string) []jsEvent {
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

func buggyLoop() {
	pollRate := time.Millisecond * 100
	devFile := "/dev/input/js0"
	state := readBuggyJoyState(devFile)

	for {
		newState := readBuggyJoyState(devFile)
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
