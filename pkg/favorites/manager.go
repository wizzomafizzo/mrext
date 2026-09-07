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

package favorites

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

const startupMarker = "# Startup favorites"

var versionedCorePattern = regexp.MustCompile(`_\d{8}\.`)

type RuntimePaths struct {
	SDRoot            string
	StartupScript     string
	ArcadeCoresFolder string
}

type FavoriteKind string

const (
	FavoriteGame       FavoriteKind = "Game"
	FavoriteCore       FavoriteKind = "Core"
	FavoriteArcadeCore FavoriteKind = "Arcade Core"
	FavoriteFolder     FavoriteKind = "Folder"
)

type Favorite struct {
	Path    string
	Target  string
	System  string
	SetName string
	Kind    FavoriteKind
}

//nolint:govet // Field order keeps configuration, paths, and caches grouped.
type Manager struct {
	cfg                 *config.UserConfig
	paths               RuntimePaths
	archiveEntriesCache map[string][]BrowseEntry
	neoGeoNamesCache    map[string]map[string]string
	createdDefaultPath  string
	createdDefaultInfo  os.FileInfo
}

func DefaultRuntimePaths() RuntimePaths {
	return RuntimePaths{
		SDRoot:            config.SdFolder,
		StartupScript:     config.StartupFile,
		ArcadeCoresFolder: config.ArcadeCoresFolder,
	}
}

func NewManager(cfg *config.UserConfig) *Manager {
	return NewManagerWithPaths(cfg, DefaultRuntimePaths())
}

func NewManagerWithPaths(cfg *config.UserConfig, paths RuntimePaths) *Manager {
	return &Manager{
		cfg:                 cfg,
		paths:               paths,
		archiveEntriesCache: make(map[string][]BrowseEntry),
		neoGeoNamesCache:    make(map[string]map[string]string),
	}
}

func (m *Manager) Root() string {
	return m.paths.SDRoot
}

func (m *Manager) RelativePath(path string) string {
	relative, err := filepath.Rel(m.paths.SDRoot, path)
	if err != nil || strings.HasPrefix(relative, "..") {
		return path
	}
	return relative
}

func (m *Manager) IsFavoriteFolderName(name string) bool {
	lowerName := strings.ToLower(name)
	if !strings.HasPrefix(name, "_") {
		return false
	}
	if strings.EqualFold(name, m.cfg.Favorites.DefaultFolder) {
		return true
	}
	for _, fragment := range m.cfg.Favorites.FolderNameContains {
		if strings.Contains(lowerName, strings.ToLower(fragment)) {
			return true
		}
	}
	return false
}

func (m *Manager) FavoriteFolders() ([]string, error) {
	entries, err := os.ReadDir(m.paths.SDRoot)
	if err != nil {
		return nil, fmt.Errorf("read MiSTer root: %w", err)
	}
	folders := make([]string, 0)
	for _, entry := range entries {
		path := filepath.Join(m.paths.SDRoot, entry.Name())
		if (entry.IsDir() || isDirectorySymlink(path, entry)) && m.IsFavoriteFolderName(entry.Name()) {
			folders = append(folders, path)
		}
	}
	sort.Slice(folders, func(i, j int) bool {
		return strings.ToLower(folders[i]) < strings.ToLower(folders[j])
	})
	return folders, nil
}

func (m *Manager) DestinationFolders(includeRoot bool, ignorePath string) ([]string, error) {
	favoritesFolders, err := m.FavoriteFolders()
	if err != nil {
		return nil, err
	}
	folders := make([]string, 0, len(favoritesFolders)+1)
	if includeRoot {
		folders = append(folders, m.paths.SDRoot)
	}
	for _, root := range favoritesFolders {
		folders = append(folders, root)
		walkErr := walkFavoriteRoot(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if path == root || !entry.IsDir() {
				return nil
			}
			if path == ignorePath {
				return filepath.SkipDir
			}
			if strings.HasPrefix(entry.Name(), "_") {
				folders = append(folders, path)
			}
			return nil
		})
		if walkErr != nil {
			return nil, fmt.Errorf("scan Favorites folders: %w", walkErr)
		}
	}
	sort.Slice(folders, func(i, j int) bool {
		return strings.ToLower(folders[i]) < strings.ToLower(folders[j])
	})
	return folders, nil
}

// Follow only the selected root symlink, not nested directory symlinks.
// Callbacks retain logical menu paths rather than exposing the target location.
func walkFavoriteRoot(root string, visit fs.WalkDirFunc) error {
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("resolve Favorites root: %w", err)
	}
	err = filepath.WalkDir(resolved, func(path string, entry fs.DirEntry, walkErr error) error {
		relative, relErr := filepath.Rel(resolved, path)
		if relErr != nil {
			return fmt.Errorf("resolve logical Favorites path: %w", relErr)
		}
		return visit(filepath.Join(root, relative), entry, walkErr)
	})
	if err != nil {
		return fmt.Errorf("walk Favorites root: %w", err)
	}
	return nil
}

func IsFavoriteFile(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))
	if extension == ".mgl" {
		return true
	}
	if extension != ".rbf" && extension != ".mra" {
		return false
	}
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}

func (m *Manager) List() ([]Favorite, error) {
	favorites := make([]Favorite, 0)
	rootEntries, err := os.ReadDir(m.paths.SDRoot)
	if err != nil {
		return nil, fmt.Errorf("read MiSTer root: %w", err)
	}
	for _, entry := range rootEntries {
		path := filepath.Join(m.paths.SDRoot, entry.Name())
		if IsFavoriteFile(path) {
			favorites = append(favorites, inspectFavorite(path))
		}
	}

	folders, err := m.FavoriteFolders()
	if err != nil {
		return nil, err
	}
	for _, folder := range folders {
		walkErr := walkFavoriteRoot(folder, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() && IsFavoriteFile(path) {
				favorites = append(favorites, inspectFavorite(path))
			}
			return nil
		})
		if walkErr != nil {
			return nil, fmt.Errorf("scan Favorites: %w", walkErr)
		}
	}
	sort.Slice(favorites, func(i, j int) bool {
		return strings.ToLower(favorites[i].Path) < strings.ToLower(favorites[j].Path)
	})
	return favorites, nil
}

func (m *Manager) EditableItems() ([]Favorite, error) {
	items, err := m.List()
	if err != nil {
		return nil, err
	}
	folders, err := m.FavoriteFolders()
	if err != nil {
		return nil, err
	}
	for _, root := range folders {
		walkErr := walkFavoriteRoot(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if path != root && entry.IsDir() && strings.HasPrefix(entry.Name(), "_") {
				items = append(items, Favorite{Path: path, Kind: FavoriteFolder})
			}
			return nil
		})
		if walkErr != nil {
			return nil, fmt.Errorf("scan editable Favorites folders: %w", walkErr)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Path) < strings.ToLower(items[j].Path)
	})
	return items, nil
}

func inspectFavorite(path string) Favorite {
	favorite := Favorite{Path: path, Target: favoriteTarget(path)}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".rbf":
		favorite.Kind = FavoriteCore
	case ".mra":
		favorite.Kind = FavoriteArcadeCore
	case ".mgl":
		favorite.Kind = FavoriteGame
		if launcher, err := mister.ReadMGL(path); err == nil {
			favorite.System = filepath.Base(launcher.Rbf)
			favorite.SetName = launcher.SetName
		}
	}
	return favorite
}

func favoriteTarget(path string) string {
	info, err := os.Lstat(path)
	if err != nil {
		return ""
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, readErr := os.Readlink(path)
		if readErr == nil {
			return target
		}
		return ""
	}
	if !strings.EqualFold(filepath.Ext(path), ".mgl") {
		return ""
	}
	launcher, err := mister.ReadMGL(path)
	if err != nil {
		return ""
	}
	const generatedPrefix = "../../../../../"
	if strings.HasPrefix(launcher.File.Path, generatedPrefix) {
		return string(filepath.Separator) + strings.TrimPrefix(launcher.File.Path, generatedPrefix)
	}
	return launcher.File.Path
}

func (m *Manager) CreateDefaultFolder() (bool, error) {
	folders, err := m.FavoriteFolders()
	if err != nil {
		return false, err
	}
	path := filepath.Join(m.paths.SDRoot, m.cfg.Favorites.DefaultFolder)
	if len(folders) > 0 {
		return false, nil
	}
	if _, statErr := os.Stat(path); statErr == nil {
		return false, nil
	} else if !os.IsNotExist(statErr) {
		return false, fmt.Errorf("inspect default Favorites folder: %w", statErr)
	}
	// #nosec G301 -- MiSTer menu folders must be readable by the menu process.
	if mkdirErr := os.Mkdir(path, 0o755); mkdirErr != nil {
		return false, fmt.Errorf("create default Favorites folder: %w", mkdirErr)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return false, fmt.Errorf("inspect created Favorites folder: %w", err)
	}
	m.createdDefaultPath = path
	m.createdDefaultInfo = info
	return true, nil
}

func (m *Manager) CleanupCreatedDefault(created bool) error {
	if !created || m.createdDefaultInfo == nil {
		return nil
	}
	path := m.createdDefaultPath
	info, statErr := os.Lstat(path)
	if os.IsNotExist(statErr) {
		return nil
	}
	if statErr != nil {
		return fmt.Errorf("inspect created Favorites folder: %w", statErr)
	}
	if !os.SameFile(info, m.createdDefaultInfo) {
		return nil
	}
	entries, err := os.ReadDir(path)
	if os.IsNotExist(err) {
		return nil //nolint:nilerr // Missing app-created folder is already clean.
	}
	if err != nil {
		return fmt.Errorf("read default Favorites folder: %w", err)
	}
	if len(entries) == 1 && entries[0].Name() == "cores" {
		coresPath := filepath.Join(path, "cores")
		info, statErr := os.Lstat(coresPath)
		if os.IsNotExist(statErr) {
			return nil //nolint:nilerr // Concurrent removal already completed cleanup.
		}
		if statErr != nil {
			return fmt.Errorf("inspect default arcade cores link: %w", statErr)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			return nil
		}
		if err := os.Remove(coresPath); err != nil {
			return fmt.Errorf("remove default arcade cores link: %w", err)
		}
		entries = nil
	}
	if len(entries) == 0 {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove unused default Favorites folder: %w", err)
		}
	}
	return nil
}

func (m *Manager) CreateFolder(parent, name string) (string, error) {
	if err := ValidateFolderName(name); err != nil {
		return "", err
	}
	if err := m.validateDestination(parent, false); err != nil {
		return "", err
	}
	path := filepath.Join(parent, name)
	if _, err := os.Lstat(path); err == nil {
		return "", fmt.Errorf("folder already exists: %s", name)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect destination folder: %w", err)
	}
	// #nosec G301 -- MiSTer menu folders must be readable by the menu process.
	if err := os.Mkdir(path, 0o755); err != nil {
		return "", fmt.Errorf("create Favorites folder: %w", err)
	}
	if m.cfg.Favorites.ManageArcadeCoreLinks {
		if err := m.ensureArcadeLink(path); err != nil {
			return "", err
		}
	}
	return path, nil
}

func ValidateFolderName(name string) error {
	switch {
	case !strings.HasPrefix(name, "_"):
		return errors.New("folder name must start with an underscore")
	case name == "_":
		return errors.New("folder name cannot be empty")
	case HasBadChars(name):
		return fmt.Errorf("folder name cannot contain any of these characters: %s", BadCharacters)
	default:
		return nil
	}
}

func (*Manager) Rename(path, name string) (string, error) {
	if info, err := os.Lstat(path); err != nil {
		return "", fmt.Errorf("inspect favorite: %w", err)
	} else if info.IsDir() {
		if err := ValidateFolderName(name); err != nil {
			return "", err
		}
	} else if err := ValidateDisplayName(name); err != nil {
		return "", err
	}
	newPath := filepath.Join(filepath.Dir(path), name)
	if !isDirectory(path) && !strings.EqualFold(filepath.Ext(name), filepath.Ext(path)) {
		newPath += filepath.Ext(path)
	}
	if _, err := os.Lstat(newPath); err == nil {
		return "", fmt.Errorf("favorite already exists: %s", filepath.Base(newPath))
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect renamed favorite: %w", err)
	}
	if err := os.Rename(path, newPath); err != nil {
		return "", fmt.Errorf("rename favorite: %w", err)
	}
	return newPath, nil
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (m *Manager) Move(path, destination string) (string, error) {
	if err := m.validateDestination(destination, !isDirectory(path)); err != nil {
		return "", err
	}
	newPath := filepath.Join(destination, filepath.Base(path))
	if _, err := os.Lstat(newPath); err == nil {
		return "", fmt.Errorf("favorite already exists in destination: %s", filepath.Base(path))
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect move destination: %w", err)
	}
	if err := os.Rename(path, newPath); err != nil {
		return "", fmt.Errorf("move favorite: %w", err)
	}
	return newPath, nil
}

func (*Manager) Delete(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect favorite for deletion: %w", err)
	}
	if !info.IsDir() {
		if !IsFavoriteFile(path) {
			return fmt.Errorf("not a managed favorite: %s", path)
		}
		if removeErr := os.Remove(path); removeErr != nil {
			return fmt.Errorf("delete favorite: %w", removeErr)
		}
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read Favorites folder: %w", err)
	}
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		entryInfo, statErr := os.Lstat(entryPath)
		if entry.Name() != "cores" || statErr != nil || entryInfo.Mode()&os.ModeSymlink == 0 {
			return errors.New("folder is not empty")
		}
	}
	for _, entry := range entries {
		if err := os.Remove(filepath.Join(path, entry.Name())); err != nil {
			return fmt.Errorf("remove arcade cores link: %w", err)
		}
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete Favorites folder: %w", err)
	}
	return nil
}

func (m *Manager) validateDestination(path string, allowRoot bool) error {
	cleanPath := filepath.Clean(path)
	if cleanPath == filepath.Clean(m.paths.SDRoot) {
		if allowRoot {
			return nil
		}
		return errors.New("top-level destination is not allowed")
	}
	folders, err := m.DestinationFolders(false, "")
	if err != nil {
		return err
	}
	for _, folder := range folders {
		if cleanPath == filepath.Clean(folder) {
			return nil
		}
	}
	return fmt.Errorf("not a Favorites folder: %s", path)
}

func (m *Manager) Refresh() error {
	favorites, err := m.List()
	if err != nil {
		return err
	}
	for _, favorite := range favorites {
		info, statErr := os.Lstat(favorite.Path)
		if statErr != nil || info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		target := favorite.Target
		resolvedTarget := target
		if !filepath.IsAbs(target) {
			resolvedTarget = filepath.Join(filepath.Dir(favorite.Path), target)
		}
		if _, statErr := os.Stat(resolvedTarget); statErr == nil {
			continue
		} else if !os.IsNotExist(statErr) {
			return fmt.Errorf("inspect favorite target: %w", statErr)
		}
		// A target on detached storage is out of reach, not deleted. Refresh
		// also runs from user-startup.sh, before USB and network mounts
		// settle, so removing here would wipe every favorite pointing at a
		// drive that is merely unplugged or a NAS that is powered off.
		if !mister.TargetAvailable(resolvedTarget) {
			continue
		}
		if err := os.Remove(favorite.Path); err != nil {
			return fmt.Errorf("remove broken favorite: %w", err)
		}
		// Users can rename a core favorite without retaining its date suffix.
		// The target identifies the installed core, not the display name.
		if !versionedCorePattern.MatchString(filepath.Base(resolvedTarget)) {
			continue
		}
		if err := refreshVersionedCore(favorite.Path, resolvedTarget); err != nil {
			return err
		}
	}
	return nil
}

func refreshVersionedCore(linkPath, oldTarget string) error {
	targetPrefix := oldTarget
	if underscore := strings.LastIndex(targetPrefix, "_"); underscore >= 0 {
		targetPrefix = targetPrefix[:underscore]
	}
	matches, err := filepath.Glob(targetPrefix + "_*")
	if err != nil {
		return fmt.Errorf("find updated core: %w", err)
	}
	if len(matches) == 0 {
		return nil
	}
	sort.Strings(matches)
	newTarget := matches[len(matches)-1]
	underscore := strings.LastIndex(newTarget, "_")
	if underscore < 0 {
		return nil
	}
	newLink := linkPath
	if versionedCorePattern.MatchString(filepath.Base(linkPath)) {
		linkPrefix := strings.TrimSuffix(linkPath, filepath.Ext(linkPath))
		linkPrefix = linkPrefix[:strings.LastIndex(linkPrefix, "_")]
		newLink = linkPrefix + newTarget[underscore:]
	}
	if err := os.Symlink(newTarget, newLink); err != nil {
		return fmt.Errorf("link updated core: %w", err)
	}
	return nil
}

func (m *Manager) TryAddToStartup() error {
	data, err := os.ReadFile(m.paths.StartupScript)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read startup script: %w", err)
	}
	if strings.Contains(string(data), startupMarker) {
		return nil
	}
	file, err := os.OpenFile(m.paths.StartupScript, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("open startup script: %w", err)
	}
	defer func() { _ = file.Close() }()
	appPath := m.cfg.AppPath
	if appPath == "" {
		appPath = "/media/fat/Scripts/favorites.sh"
	}
	commandPath := appPath
	if strings.ContainsAny(appPath, " \t") {
		commandPath = strconv.Quote(appPath)
	}
	entry := fmt.Sprintf("\n# Startup favorites\n[[ -e %s ]] && %s refresh\n", commandPath, commandPath)
	if _, err := file.WriteString(entry); err != nil {
		return fmt.Errorf("append Favorites startup entry: %w", err)
	}
	return nil
}

func (m *Manager) SetupArcadeLinks() error {
	if !m.cfg.Favorites.ManageArcadeCoreLinks {
		return nil
	}
	if err := m.ensureArcadeLink(m.paths.SDRoot); err != nil {
		return err
	}
	folders, err := m.FavoriteFolders()
	if err != nil {
		return err
	}
	for _, root := range folders {
		walkErr := walkFavoriteRoot(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return m.ensureArcadeLink(path)
			}
			return nil
		})
		if walkErr != nil {
			return fmt.Errorf("set up arcade links: %w", walkErr)
		}
	}
	return nil
}

func (m *Manager) ensureArcadeLink(folder string) error {
	linkPath := filepath.Join(folder, "cores")
	if _, err := os.Lstat(linkPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect arcade cores link: %w", err)
	}
	if err := os.Symlink(m.paths.ArcadeCoresFolder, linkPath); err != nil {
		return fmt.Errorf("create arcade cores link: %w", err)
	}
	return nil
}
