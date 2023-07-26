package mister

import (
	"github.com/wizzomafizzo/mrext/pkg/config"
	"net"
	"os"
	"regexp"
	"strings"
)

func readUBootConfig() (string, error) {
	uBootConfigData, err := os.ReadFile(config.UBootConfigFile)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	data := string(uBootConfigData)
	data = strings.ReplaceAll(data, "\n", " ")
	data = strings.ReplaceAll(data, "\r", " ")

	return data, nil
}

var ethAddrArg = regexp.MustCompile(`ethaddr=([0-9a-fA-F]{2}(:[0-9a-fA-F]{2}){5})`)

// GetConfiguredMacAddress returns the ethernet MAC address configured in the u-boot.txt file, if available.
func GetConfiguredMacAddress() (string, error) {
	uBootConfig, err := readUBootConfig()
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(uBootConfig, "\n") {
		if ethAddrArg.MatchString(line) {
			return ethAddrArg.FindStringSubmatch(line)[1], nil
		}
	}

	return "", nil
}

// UpdateConfiguredMacAddress updates the ethernet MAC address configured in the u-boot.txt file. Setting a new one if
// it doesn't exist, or updating the existing one. Any existing u-boot.txt arguments are preserved.
func UpdateConfiguredMacAddress(newMacAddress string) error {
	uBootConfig, err := readUBootConfig()
	if err != nil {
		return err
	}

	uBootConfig = ethAddrArg.ReplaceAllString(uBootConfig, "")

	if newMacAddress != "" {
		_, err = net.ParseMAC(newMacAddress)
		if err != nil {
			return err
		}

		uBootConfig += " ethaddr=" + newMacAddress
	}

	uBootConfig = strings.TrimSpace(uBootConfig)

	return os.WriteFile(config.UBootConfigFile, []byte(uBootConfig), 0644)
}

var usbhidQuirksArg = regexp.MustCompile(`usbhid\.quirks=([0-9a-fA-F:,xXuU]+) *`)

func GetUsbHidQuirks() ([]string, error) {
	var quirks []string

	uBootConfig, err := readUBootConfig()
	if err != nil {
		return quirks, err
	}

	if uBootConfig == "" {
		return quirks, nil
	}

	for _, line := range strings.Split(uBootConfig, "\n") {
		if usbhidQuirksArg.MatchString(line) {
			match := usbhidQuirksArg.FindStringSubmatch(line)[1]
			quirks = strings.Split(match, ",")
			break
		}
	}

	return quirks, nil
}

func UpdateUsbHidQuirks(quirks []string) error {
	uBootConfig, err := readUBootConfig()
	if err != nil {
		return err
	}

	uBootConfig = usbhidQuirksArg.ReplaceAllString(uBootConfig, "")

	if len(quirks) > 0 {
		uBootConfig += " usbhid.quirks=" + strings.Join(quirks, ",")
	}

	uBootConfig = strings.TrimSpace(uBootConfig)

	return os.WriteFile(config.UBootConfigFile, []byte(uBootConfig), 0644)
}
