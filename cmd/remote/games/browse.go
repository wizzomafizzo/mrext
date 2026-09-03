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

package games

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/cmd/remote/menu"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type FolderResult struct {
	Path   string       `json:"path"`
	System games.System `json:"system"`
}

func getGamesFolders() []FolderResult {
	systemResults := make(map[string]FolderResult)
	folderNames := make(map[string]games.System)

	for id := range games.Systems {
		system := games.Systems[id]
		folder := strings.ToLower(system.Folder[0])
		folderNames[folder] = system
	}

	for _, root := range config.GamesFolders {
		if _, err := os.Stat(root); err != nil {
			continue
		}

		gfs, err := os.ReadDir(root)
		if err != nil {
			continue
		}

		for _, gf := range gfs {
			if !gf.IsDir() {
				continue
			}

			folder := strings.ToLower(gf.Name())

			if _, ok := folderNames[folder]; !ok {
				continue
			}

			system := folderNames[folder]

			_, ok := systemResults[system.Id]
			if ok {
				continue
			}

			systemResults[system.Id] = FolderResult{
				System: system,
				Path:   filepath.Join(root, gf.Name()),
			}
		}
	}

	folders := make([]FolderResult, 0, len(systemResults))
	for id := range systemResults {
		folders = append(folders, systemResults[id])
	}

	return folders
}

type fileEntry struct {
	modTime time.Time
	path    string
	name    string
	size    int64
	isDir   bool
}

func listPath(logger *service.Logger, path string) ([]menu.Item, error) {
	systems := games.FolderToSystems(&config.UserConfig{}, path+"/")
	logger.Info("systems: %v", systems)

	inZip := false
	zipIndex := -1
	parts := strings.Split(path, "/")

	for i, part := range parts {
		if strings.HasSuffix(strings.ToLower(part), ".zip") {
			inZip = true
			zipIndex = i
			break
		}
	}

	files := make([]fileEntry, 0)

	if inZip {
		zipFile := strings.Join(parts[:zipIndex+1], "/")
		zipPath := strings.Join(parts[zipIndex+1:], "/")

		paths, err := utils.ListZip(zipFile)
		if err != nil {
			return nil, fmt.Errorf("list ZIP contents: %w", err)
		}

		if len(paths) == 0 {
			return nil, nil
		}

		for _, zipItem := range paths {
			if !strings.HasPrefix(zipItem, zipPath) {
				continue
			}

			isDir := false
			if strings.HasSuffix(zipItem, "/") {
				isDir = true
				zipItem = strings.TrimSuffix(zipItem, "/")
			}

			if zipItem == zipPath {
				continue
			}

			if strings.Count(strings.TrimPrefix(zipItem, zipPath+"/"), "/") != 0 {
				continue
			}

			files = append(files, fileEntry{
				path:    filepath.Join(zipFile, zipItem),
				name:    filepath.Base(zipItem),
				isDir:   isDir,
				modTime: time.Time{},
			})
		}
	} else {
		fsFiles, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("read games directory: %w", err)
		}

		for _, fsFile := range fsFiles {
			info, err := fsFile.Info()
			if err != nil {
				continue
			}

			files = append(files, fileEntry{
				path:    filepath.Join(path, fsFile.Name()),
				name:    fsFile.Name(),
				size:    info.Size(),
				isDir:   fsFile.IsDir(),
				modTime: info.ModTime(),
			})
		}
	}

	validFiletypes := make([]string, 0)

	if !inZip {
		validFiletypes = append(validFiletypes, ".zip")
	}

	for systemIndex := range systems {
		system := &systems[systemIndex]
		for _, slot := range system.Slots {
			validFiletypes = append(validFiletypes, slot.Exts...)
		}
	}
	logger.Info("valid filetypes: %s", validFiletypes)

	items := make([]menu.Item, 0)

	for fileIndex := range files {
		file := &files[fileIndex]
		var friendlyName string
		if file.isDir {
			friendlyName = file.name
		} else {
			friendlyName = strings.TrimSuffix(file.name, filepath.Ext(file.name))
		}

		if strings.HasPrefix(file.name, ".") {
			continue
		}

		if !file.isDir && !utils.ContainsFold(validFiletypes, filepath.Ext(file.name)) {
			continue
		}

		var next *string
		filetype := "game"
		var system *menu.MenuSystem

		if file.isDir {
			nextPath := filepath.Join(path, file.name)
			next = &nextPath
			filetype = "folder"
		} else {
			match, err := games.BestSystemMatch(&config.UserConfig{}, file.path)
			if err == nil {
				system = &menu.MenuSystem{
					ID:       match.Id,
					Name:     match.Name,
					Category: match.Category,
				}
			}
		}

		if strings.ToLower(filepath.Ext(file.name)) == ".zip" {
			nextPath := filepath.Join(path, file.name)
			next = &nextPath
			filetype = "zip"
		}

		items = append(items, menu.Item{
			Name:      friendlyName,
			Path:      filepath.Join(path, file.name),
			Parent:    path,
			Filename:  filepath.Base(file.name),
			Extension: filepath.Ext(file.name),
			Next:      next,
			Modified:  file.modTime,
			Size:      file.size,
			Type:      filetype,
			InZip:     inZip,
			System:    system,
		})
	}

	return items, nil
}

type ListGamesPayload struct {
	Up    *string     `json:"up,omitempty"`
	Items []menu.Item `json:"items"`
	// TODO: system
}

func ListGamesFolder(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("list games folder")

		var args struct {
			Path string `json:"path"`
		}

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Error("error decoding request: %s", err)
			return
		}

		items := make([]menu.Item, 0)
		var up *string

		systemFolders := getGamesFolders()
		systemFoldersMap := make(map[string]bool)
		for i := range systemFolders {
			folder := &systemFolders[i]
			systemFoldersMap[strings.ToLower(folder.Path)] = true
		}

		// list system folders instead
		if args.Path == "" {
			up = nil
			for i := range systemFolders {
				folder := &systemFolders[i]
				var next *string
				nextPath := folder.Path
				next = &nextPath

				items = append(items, menu.Item{
					Name:      filepath.Base(folder.Path),
					Path:      folder.Path,
					Parent:    args.Path,
					Filename:  filepath.Base(folder.Path),
					Extension: filepath.Ext(folder.Path),
					Next:      next,
					Type:      "folder",
				})
			}

			err = json.NewEncoder(w).Encode(ListGamesPayload{
				Up:    up,
				Items: items,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.Error("error encoding payload: %s", err)
				return
			}

			return
		}

		path, err := filepath.Abs(args.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Error("error getting absolute path: %s", err)
			return
		}

		valid := false
		atRoot := false
		for _, folder := range config.GamesFolders {
			if strings.EqualFold(path, folder) {
				valid = false
				break
			}

			if strings.HasPrefix(strings.ToLower(path), folder) {
				valid = true
			}

			if _, ok := systemFoldersMap[strings.ToLower(path)]; ok {
				valid = true
				atRoot = true
				break
			}
		}

		if !valid {
			http.Error(w, "invalid path", http.StatusInternalServerError)
			logger.Error("invalid path: %s", path)
			return
		}

		if atRoot {
			home := ""
			up = &home
		} else {
			upPath := filepath.Dir(path)
			up = &upPath
		}

		items, err = listPath(logger, path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error listing path: %s", err)
			return
		}

		payload := ListGamesPayload{
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
