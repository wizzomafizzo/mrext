package mister

import (
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"os"
	"path/filepath"
	"syscall"
)

func mapSharedMem(address int64) (*[]byte, *os.File, error) {
	file, err := os.OpenFile(
		"/dev/mem",
		os.O_RDWR|os.O_SYNC,
		0,
	)
	if err != nil {
		return &[]byte{}, nil, fmt.Errorf("error opening /dev/mem: %s", err)
	}

	mem, err := syscall.Mmap(
		int(file.Fd()),
		address,
		0x1000,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return &[]byte{}, nil, fmt.Errorf("error mapping /dev/mem: %s", err)
	}

	return &mem, file, nil
}

func unmapSharedMem(mem *[]byte, file *os.File) error {
	err := syscall.Munmap(*mem)
	if err != nil {
		return fmt.Errorf("error unmapping /dev/mem: %s", err)
	}

	if file == nil {
		return fmt.Errorf("/dev/mem file reference is nil")
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("error closing /dev/mem: %s", err)
	}

	return nil
}

func GetCurrentIni() (int, error) {
	mem, file, err := mapSharedMem(0x1FFFF000)
	if err != nil {
		return 0, err
	}

	offset := 0xF04
	vs := []byte{(*mem)[offset], (*mem)[offset+1], (*mem)[offset+2], (*mem)[offset+3]}

	err = unmapSharedMem(mem, file)
	if err != nil {
		return 0, err
	}

	if vs[0] == 0x34 && vs[1] == 0x99 && vs[2] == 0xBA {
		return int(vs[3] + 1), nil
	} else {
		return 0, nil
	}
}

func SetCurrentIni(ini int) error {
	if ini < 1 || ini > 4 {
		return fmt.Errorf("ini number out of range: %d", ini)
	}

	mem, file, err := mapSharedMem(0x1FFFF000)
	if err != nil {
		return err
	}

	offset := 0xF04
	(*mem)[offset] = 0x34
	(*mem)[offset+1] = 0x99
	(*mem)[offset+2] = 0xBA
	(*mem)[offset+3] = byte(ini - 1)

	err = unmapSharedMem(mem, file)
	if err != nil {
		return err
	}

	coreName, err := GetActiveCoreName()
	if err != nil {
		return err
	}

	if coreName == config.MenuCore {
		err = LaunchMenu()
		if err != nil {
			return err
		} else {
			return nil
		}
	}

	recent, err := ReadRecent(config.CoresRecentFile)
	if err != nil || len(recent) == 0 {
		err = LaunchMenu()
		if err != nil {
			return err
		} else {
			return nil
		}
	} else {
		corePath := filepath.Join(config.SdFolder, recent[0].Directory, recent[0].Name)
		err = LaunchGenericFile(corePath)
		if err != nil {
			return err
		}
	}

	return nil
}
