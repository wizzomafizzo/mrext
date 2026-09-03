package mister

import (
	"fmt"
	"os"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

const UBootMACParam = "ethaddr"

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
