// mrext
// Copyright (c) 2026 mrext contributors.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of mrext.
//
// mrext is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// mrext is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with mrext. If not, see <http://www.gnu.org/licenses/>.

package menu

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

// TODO: should be in config
const namesTxtPath = "/media/fat/names.txt"

type MenuSystem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type Item struct {
	Modified  time.Time   `json:"modified"`
	Version   *time.Time  `json:"version,omitempty"`
	Next      *string     `json:"next,omitempty"`
	NamesTxt  *string     `json:"namesTxt,omitempty"`
	System    *MenuSystem `json:"system,omitempty"`
	Path      string      `json:"path"`
	Parent    string      `json:"parent"`
	Filename  string      `json:"filename"`
	Extension string      `json:"extension"`
	Type      string      `json:"type"`
	Name      string      `json:"name"`
	Size      int64       `json:"size"`
	InZip     bool        `json:"inZip"`
}

type ListMenuPayload struct {
	Up    *string `json:"up,omitempty"`
	Items []Item  `json:"items"`
}

var namesMapping = map[string]string{}

func GetNamesTxt(original, filetype string) (string, error) {
	if filetype == "folder" {
		return "", nil
	}

	if len(namesMapping) == 0 {
		rootFS := os.DirFS("/")
		err := loadNamesMapping(rootFS)
		if err != nil {
			return "", err
		}
	}

	return namesMapping[original], nil
}

func loadNamesMapping(f fs.FS) error {
	file, err := f.Open(strings.TrimPrefix(namesTxtPath, "/"))
	if err != nil {
		return fmt.Errorf("open names mapping: %w", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) > 1 {
			namesMapping[parts[0]] = strings.Trim(parts[1], " ")
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan names mapping: %w", err)
	}
	return nil
}

func isValidMenuFile(file os.DirEntry, includeHidden bool) bool {
	name := file.Name()
	lower := strings.ToLower(name)

	if lower == "menu.rbf" {
		return false
	}

	if file.IsDir() {
		if name == "." || name == ".." {
			return false
		}

		if strings.HasPrefix(lower, "_") {
			return true
		}

		if includeHidden && strings.HasPrefix(lower, "._") {
			return true
		}
	}

	if strings.HasSuffix(lower, ".mra") || strings.HasSuffix(lower, ".rbf") || strings.HasSuffix(lower, ".mgl") {
		if !includeHidden && strings.HasPrefix(lower, ".") {
			return false
		}

		return true
	}

	return false
}

func getFileType(file os.DirEntry) string {
	name := file.Name()
	lower := strings.ToLower(name)

	if file.IsDir() {
		return "folder"
	}

	if strings.HasSuffix(lower, ".mra") {
		return "mra"
	}

	if strings.HasSuffix(lower, ".rbf") {
		return "rbf"
	}

	if strings.HasSuffix(lower, ".mgl") {
		return "mgl"
	}

	if strings.HasSuffix(lower, ".zip") {
		return "zip"
	}

	return "unknown"
}

func GetFilenameInfo(file os.DirEntry) (name, filetype string, version *time.Time) {
	name = file.Name()
	filetype = getFileType(file)

	name = strings.TrimSuffix(name, filepath.Ext(name))

	if filetype == "folder" {
		name = strings.TrimPrefix(name, "_")
		return name, filetype, nil
	}

	parts := strings.Split(name, "_")
	if len(parts) > 1 {
		ver, err := time.Parse("20060102", parts[len(parts)-1])
		if err == nil {
			version = &ver
		}

		name = strings.Join(parts[:len(parts)-1], "_")
	}

	return name, filetype, version
}

var removeRoot = regexp.MustCompile(`(?i)^` + config.SdFolder + `/?`)

func ListFolder(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("list menu folder")

		var args struct {
			Path string `json:"path"`
		}

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Error("error decoding request: %s", err)
			return
		}

		args.Path = removeRoot.ReplaceAllString(args.Path, "")

		var path string
		if args.Path == "" {
			path = config.SdFolder
		} else {
			parts := filepath.SplitList(args.Path)
			cleaned := make([]string, 0)
			cleaned = append(cleaned, config.SdFolder)

			for _, part := range parts {
				if part == "." || part == ".." {
					continue
				}

				cleaned = append(cleaned, part)
			}

			path = filepath.Join(cleaned...)
		}

		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			http.Error(w, statErr.Error(), http.StatusNotFound)
			logger.Error("menu folder (%s) does not exist: %s", path, statErr)
			return
		}

		files, err := os.ReadDir(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't list menu folder (%s): %s", path, err)
			return
		}

		items := make([]Item, 0)
		for _, file := range files {
			name := file.Name()

			formatted, filetype, version := GetFilenameInfo(file)

			info, infoErr := file.Info()
			if infoErr != nil {
				logger.Error("couldn't get file info for %s: %s", name, infoErr)
				continue
			}

			namesTxtResult, namesErr := GetNamesTxt(formatted, filetype)
			if namesErr != nil {
				logger.Error("couldn't get names.txt for %s: %s", name, namesErr)
			}

			var namesTxt *string
			if namesTxtResult != "" {
				namesTxt = &namesTxtResult
			}

			var next *string
			if file.IsDir() {
				nextPath := filepath.Join(args.Path, name)
				next = &nextPath
			}

			if isValidMenuFile(file, false) {
				items = append(items, Item{
					Name:      formatted,
					NamesTxt:  namesTxt,
					Path:      filepath.Join(path, name),
					Parent:    args.Path,
					Filename:  filepath.Base(name),
					Extension: filepath.Ext(name),
					Next:      next,
					Type:      filetype,
					Modified:  info.ModTime(),
					Version:   version,
					Size:      info.Size(),
				})
			}
		}

		var up *string
		if args.Path != "" && args.Path != "." {
			upPath := filepath.Dir(args.Path)
			up = &upPath
		}

		payload := ListMenuPayload{
			Up:    up,
			Items: items,
		}
		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error encoding payload: %s", err)
			return
		}
	}
}
