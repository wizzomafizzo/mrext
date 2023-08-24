package mister

import (
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"os"
	"regexp"
	"strings"
)

const (
	UBootMACParam    = "ethaddr"
	UBootKernelParam = "v"
)

func ReadUBootParams() (map[string]string, error) {
	params := make(map[string]string)

	data, err := os.ReadFile(config.UBootConfigFile)
	if os.IsNotExist(err) {
		return params, nil
	} else if err != nil {
		return params, err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		line = strings.TrimSpace(line)

		if line == "" || !strings.Contains(line, "=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)

		key := parts[0]
		key = strings.TrimSpace(key)

		value := parts[1]
		value = strings.TrimSpace(value)

		params[parts[0]] = parts[1]
	}

	return params, nil
}

func WriteUBootParams(params map[string]string) error {
	var pairs []string

	for key, value := range params {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}

	content := strings.Join(pairs, "\n") + "\n"

	if _, err := os.Stat(config.UBootConfigFile); err == nil {
		err = os.Rename(config.UBootConfigFile, config.UBootConfigFile+".backup")
		if err != nil {
			return err
		}
	}

	err := os.WriteFile(config.UBootConfigFile, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}

func parseKernelArgs(input string) map[string]string {
	args := make(map[string]string)

	re := regexp.MustCompile(`([\w_\-.]+)="(.*?)"|([\w_\-.]+)=(\S+)`)
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		param := match[1] + match[3]
		value := match[2] + match[4]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		args[param] = value
	}

	return args
}

func makeKernelArgs(params map[string]string) string {
	var pairs []string

	for key, value := range params {
		// if the value contains spaces, quote it
		if strings.Contains(value, " ") {
			value = fmt.Sprintf("\"%s\"", value)
		}

		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}

	content := strings.Join(pairs, " ")

	return content
}

// GetConfiguredMacAddress returns the ethernet MAC address configured in the u-boot.txt file, if available.
func GetConfiguredMacAddress() (string, error) {
	params, err := ReadUBootParams()
	if err != nil {
		return "", err
	}

	if ethAddr, ok := params[UBootMACParam]; ok {
		return ethAddr, nil
	}

	return "", nil
}

// UpdateConfiguredMacAddress updates the ethernet MAC address configured in the u-boot.txt file. Setting a new one if
// it doesn't exist, or updating the existing one. Any existing u-boot.txt arguments are preserved.
func UpdateConfiguredMacAddress(newMacAddress string) error {
	params, err := ReadUBootParams()
	if err != nil {
		return err
	}

	params[UBootMACParam] = newMacAddress

	return WriteUBootParams(params)
}

func GetUsbHidQuirks() ([]string, error) {
	params, err := ReadUBootParams()
	if err != nil {
		return nil, err
	}

	args := make(map[string]string)
	if v, ok := params[UBootKernelParam]; ok {
		args = parseKernelArgs(v)
	}

	if v, ok := args["usbhid.quirks"]; ok {
		return strings.Split(v, ","), nil
	}

	return nil, nil
}

func UpdateUsbHidQuirks(quirks []string) error {
	params, err := ReadUBootParams()
	if err != nil {
		return err
	}

	args := make(map[string]string)
	if v, ok := params[UBootKernelParam]; ok {
		args = parseKernelArgs(v)
	}

	args["usbhid.quirks"] = strings.Join(quirks, ",")
	params[UBootKernelParam] = makeKernelArgs(args)

	return WriteUBootParams(params)
}

func EnableFastUsbPoll() error {
	params, err := ReadUBootParams()
	if err != nil {
		return err
	}

	args := make(map[string]string)
	if v, ok := params[UBootKernelParam]; ok {
		args = parseKernelArgs(v)
	}

	args["loglevel"] = "4"
	args["usbhid.jspoll"] = "1"
	args["xpad.cpoll"] = "1"

	params[UBootKernelParam] = makeKernelArgs(args)

	return WriteUBootParams(params)
}

func IsFastUsbPollActive() (bool, error) {
	params, err := ReadUBootParams()
	if err != nil {
		return false, err
	}

	args := make(map[string]string)
	if v, ok := params[UBootKernelParam]; ok {
		args = parseKernelArgs(v)
	}

	if _, ok := args["usbhid.jspoll"]; ok {
		return true, nil
	}

	return false, nil
}
