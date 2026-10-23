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

package tracker

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/games"
)

// arcadeRebuildInterval is the shortest gap between index builds a missed
// lookup asks for. MiSTer echoes every arcade launch through ACTIVEGAME, and
// some sets are carried by several files, so a miss is common and must not
// turn into a rebuild loop.
const arcadeRebuildInterval = time.Minute

// arcadeIndex maps each set name found in the installed .mra files to the
// files that carry it. It is built off the tracker lock and replaced whole,
// never changed once installed.
type arcadeIndex struct {
	startedAt time.Time
	// Keys are lower-cased, trimmed set names; values are distinct paths.
	bySet map[string][]string
	stats arcadeIndexStats
}

type arcadeIndexStats struct {
	took          time.Duration
	dirs          int
	files         int
	linksSkipped  int
	linksFollowed int
}

// Replaced in tests, to count how often links are resolved.
var (
	readLink     = os.Readlink
	evalSymlinks = filepath.EvalSymlinks
)

func arcadeSetKey(setName string) string {
	return strings.ToLower(strings.TrimSpace(setName))
}

// paths returns the files carrying setName.
func (idx *arcadeIndex) paths(setName string) []string {
	return idx.bySet[arcadeSetKey(setName)]
}

type indexRoot struct {
	raw       string
	canonical string
}

type indexBuilder struct {
	idx   *arcadeIndex
	seen  map[string]bool
	roots []indexRoot
}

// buildArcadeIndex walks the arcade roots once and reads the set name of
// every .mra in them. Directory entries are classified from ReadDir alone:
// links are resolved only when their target lies outside the roots, so the
// thousands of links an _Organized tree holds cost one Readlink each and
// their files are indexed once, under their real path.
func buildArcadeIndex(roots []string) *arcadeIndex {
	started := time.Now()
	b := &indexBuilder{
		idx:  &arcadeIndex{startedAt: started, bySet: make(map[string][]string)},
		seen: make(map[string]bool),
	}
	b.roots = indexRoots(roots)
	for _, root := range b.roots {
		b.seen[root.canonical] = true
	}
	for _, root := range b.roots {
		b.visitDir(root.raw, false)
	}
	b.idx.stats.took = time.Since(started)
	return b.idx
}

// indexRoots cleans the roots and drops those that do not resolve, repeat
// another, or lie inside another, so no tree is walked twice.
func indexRoots(roots []string) []indexRoot {
	var resolved []indexRoot
	for _, root := range roots {
		if root == "" {
			continue
		}
		raw := filepath.Clean(root)
		canonical, err := evalSymlinks(raw)
		if err != nil {
			continue
		}
		resolved = append(resolved, indexRoot{raw: raw, canonical: canonical})
	}
	var kept []indexRoot
	for i, root := range resolved {
		inside := false
		for j, other := range resolved {
			if i == j {
				continue
			}
			if root.canonical == other.canonical && j < i ||
				root.canonical != other.canonical && within(root.canonical, other.canonical) {
				inside = true
				break
			}
		}
		if !inside {
			kept = append(kept, root)
		}
	}
	return kept
}

// within reports whether path is dir or lies under it.
func within(path, dir string) bool {
	if path == dir {
		return true
	}
	return strings.HasPrefix(path, strings.TrimSuffix(dir, string(filepath.Separator))+string(filepath.Separator))
}

// insideRoots reports whether path lies in a root, in its raw or its
// canonical form.
func (b *indexBuilder) insideRoots(path string) bool {
	for _, root := range b.roots {
		if within(path, root.raw) || within(path, root.canonical) {
			return true
		}
	}
	return false
}

// visitDir indexes a directory. followed is set below a link that leads
// outside the roots, where real entries are checked against seen too, so
// a tree reached through more than one link is indexed once.
func (b *indexBuilder) visitDir(dir string, followed bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	b.idx.stats.dirs++
	for _, entry := range entries {
		name := entry.Name()
		// AppleDouble files a Mac leaves beside the real ones.
		if strings.HasPrefix(name, "._") {
			continue
		}
		path := filepath.Join(dir, name)
		mode := entry.Type()
		switch {
		case mode&fs.ModeSymlink != 0:
			b.visitLink(dir, path)
		case mode.IsDir():
			if !followed || b.firstVisit(path) {
				b.visitDir(path, followed)
			}
		case mode.IsRegular() && isMRA(name):
			if !followed || b.firstVisit(path) {
				b.add(path)
			}
		}
	}
}

func (b *indexBuilder) visitLink(dir, path string) {
	target, err := readLink(path)
	if err != nil {
		return
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(dir, target)
	}
	if b.insideRoots(filepath.Clean(target)) {
		// The walk reaches the target itself.
		b.idx.stats.linksSkipped++
		return
	}
	canonical, err := evalSymlinks(path)
	if err != nil || b.seen[canonical] || b.insideRoots(canonical) {
		b.idx.stats.linksSkipped++
		return
	}
	b.seen[canonical] = true
	info, err := os.Stat(canonical)
	if err != nil {
		return
	}
	b.idx.stats.linksFollowed++
	switch {
	case info.IsDir():
		b.visitDir(path, true)
	case info.Mode().IsRegular() && isMRA(path):
		b.add(path)
	}
}

// firstVisit marks an entry below a followed link as seen, and reports
// whether it was new.
func (b *indexBuilder) firstVisit(path string) bool {
	canonical, err := evalSymlinks(path)
	if err != nil || b.seen[canonical] {
		return false
	}
	b.seen[canonical] = true
	return true
}

func (b *indexBuilder) add(path string) {
	b.idx.stats.files++
	key := arcadeSetKey(readMRASetName(path))
	if key == "" {
		return
	}
	b.idx.bySet[key] = append(b.idx.bySet[key], path)
}

func isMRA(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".mra")
}

// arcadeRootsNow lists the folders the arcade index walks.
func (tr *Tracker) arcadeRootsNow() []string {
	if tr.arcadeRoots != nil {
		return tr.arcadeRoots()
	}
	system, err := games.GetSystem(ArcadeSystem)
	if err != nil {
		return nil
	}
	paths := games.GetSystemPaths(tr.Config, []games.System{*system})
	roots := make([]string, 0, len(paths))
	for i := range paths {
		roots = append(roots, paths[i].Path)
	}
	return roots
}

// requestArcadeIndex starts an index build unless one is running. It runs
// under the tracker lock; the build does not, so a full SD card never holds
// up core and game changes.
func (tr *Tracker) requestArcadeIndex() {
	if tr.indexBusy.Load() {
		return
	}
	tr.indexBusy.Store(true)
	tr.lastBuildStart = time.Now()
	run := tr.runBuild
	if run == nil {
		run = func(build func()) { go build() }
	}
	run(func() {
		tr.installArcadeIndex(buildArcadeIndex(tr.arcadeRootsNow()))
	})
}

// warmArcadeIndex starts the first index build, if nothing has yet. It runs
// under the tracker lock.
func (tr *Tracker) warmArcadeIndex() {
	if tr.arcadeIdx.Load() == nil {
		tr.requestArcadeIndex()
	}
}

// installArcadeIndex replaces the index and brings the running game up to
// date with it in one critical section, so no core or game change can land
// in between.
func (tr *Tracker) installArcadeIndex(idx *arcadeIndex) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.arcadeIdx.Store(idx)
	tr.indexBusy.Store(false)
	if tr.Logger != nil {
		tr.Logger.Info("indexed %d arcade sets from %d MRA files in %s (%d folders, %d links skipped, %d followed)",
			len(idx.bySet), idx.stats.files, idx.stats.took.Round(time.Millisecond), idx.stats.dirs,
			idx.stats.linksSkipped, idx.stats.linksFollowed)
	}
	tr.reconcileArcade()
	tr.updateSessions()
	if ready := tr.indexReady(); !tr.indexInstalled {
		tr.indexInstalled = true
		close(ready)
	}
}

// indexReady returns the channel closed once the first index is installed.
// It runs under the tracker lock.
func (tr *Tracker) indexReady() chan struct{} {
	if tr.indexDone == nil {
		tr.indexDone = make(chan struct{})
	}
	return tr.indexDone
}

// reconcileArcade brings the running game up to date with a new index. It
// runs under the tracker lock and asks for no further build. An arcade game
// that started before the index could name its file gets its path, with no
// event. An unlisted set that came up before the first index is taken up
// as an arcade game if its core still runs with nothing else tracked, and
// sets the new index finds are no longer taken for unknown.
func (tr *Tracker) reconcileArcade() {
	id := tr.ActiveCore
	if id != "" && tr.ActiveSystem == ArcadeSystem && tr.ActiveGame == id && tr.ActiveGamePath == "" {
		if path, _ := tr.findArcadeSetPath(id, false); path != "" {
			tr.setArcadePath(id, path)
		}
	}

	for name := range tr.unknownSets {
		if path, _ := tr.findArcadeSetPath(name, false); path != "" {
			delete(tr.unknownSets, name)
		}
	}

	name := tr.pendingSet
	if name == "" {
		return
	}
	tr.pendingSet = ""
	path, _ := tr.findArcadeSetPath(name, false)
	if path == "" {
		tr.markUnknownSet(name)
		return
	}
	if name == tr.ActiveCore && tr.ActiveSystem == "" && tr.ActiveGame == "" && tr.coreStillRunning(name) {
		tr.beginArcade(unlistedArcade(name, path))
	}
}

// WaitArcadeIndex starts the arcade index if needed and waits up to timeout
// for the first one, reporting whether it is ready. For a caller that must
// have arcade paths from its first core change on.
func (tr *Tracker) WaitArcadeIndex(timeout time.Duration) bool {
	tr.mu.Lock()
	tr.warmArcadeIndex()
	ready := tr.indexReady()
	tr.mu.Unlock()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-ready:
		return true
	case <-timer.C:
		return false
	}
}
