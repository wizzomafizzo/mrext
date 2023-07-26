package mister

import (
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"os"
	"regexp"
	"strings"
)

func ReadUBootParams() (map[string]string, error) {
	params := make(map[string]string)

	data, err := os.ReadFile(config.UBootConfigFile)
	if os.IsNotExist(err) {
		return params, nil
	} else if err != nil {
		return params, err
	}

	input := string(data)
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")

	re := regexp.MustCompile(`(\w+)="(.*?)"|(\w+)=(\S+)`)
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		param := match[1] + match[3]
		value := match[2] + match[4]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		params[param] = value
	}

	return params, nil
}

func WriteUBootParams(params map[string]string) error {
	var pairs []string

	for key, value := range params {
		// if the value contains spaces, quote it
		if strings.Contains(value, " ") {
			value = fmt.Sprintf("\"%s\"", value)
		}

		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}

	content := strings.Join(pairs, " ")

	err := os.WriteFile(config.UBootConfigFile, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}

// GetConfiguredMacAddress returns the ethernet MAC address configured in the u-boot.txt file, if available.
func GetConfiguredMacAddress() (string, error) {
	params, err := ReadUBootParams()
	if err != nil {
		return "", err
	}

	if ethAddr, ok := params["ethaddr"]; ok {
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

	params["ethaddr"] = newMacAddress

	return WriteUBootParams(params)
}

func GetUsbHidQuirks() ([]string, error) {
	params, err := ReadUBootParams()
	if err != nil {
		return nil, err
	}

	if quirks, ok := params["usbhid.quirks"]; ok {
		return strings.Split(quirks, ","), nil
	}

	return nil, nil
}

func UpdateUsbHidQuirks(quirks []string) error {
	params, err := ReadUBootParams()
	if err != nil {
		return err
	}

	params["usbhid.quirks"] = strings.Join(quirks, ",")

	return WriteUBootParams(params)
}
