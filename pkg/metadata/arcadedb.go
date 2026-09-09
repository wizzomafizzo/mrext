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

package metadata

import (
	"context"
	// #nosec G505 -- Git object hashing is defined as SHA-1. Used to compare a
	// downloaded file against the hash GitHub reports, not for security.
	"crypto/sha1" //nolint:gosec // see above
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/wizzomafizzo/mrext/pkg/config"
)

type gitHubLinks struct {
	Self string `json:"self"`
	Git  string `json:"git"`
	HTML string `json:"html"`
}

type gitHubContentsItem struct {
	Links       gitHubLinks `json:"_links"` //nolint:tagliatelle // GitHub API field name.
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	SHA         string      `json:"sha"`
	URL         string      `json:"url"`
	HTMLURL     string      `json:"html_url"`     //nolint:tagliatelle // GitHub API field name.
	GitURL      string      `json:"git_url"`      //nolint:tagliatelle // GitHub API field name.
	DownloadURL string      `json:"download_url"` //nolint:tagliatelle // GitHub API field name.
	Type        string      `json:"type"`
	Size        int         `json:"size"`
}

type ArcadeDBEntry struct {
	Setname         string `csv:"setname"`
	Name            string `csv:"name"`
	Region          string `csv:"region"`
	Version         string `csv:"version"`
	Alternative     string `csv:"alternative"`
	ParentTitle     string `csv:"parent_title"`
	Platform        string `csv:"platform"`
	Series          string `csv:"series"`
	Homebrew        string `csv:"homebrew"`
	Bootleg         string `csv:"bootleg"`
	Year            string `csv:"year"`
	Manufacturer    string `csv:"manufacturer"`
	Category        string `csv:"category"`
	Linebreak1      string `csv:"linebreak1"`
	Resolution      string `csv:"resolution"`
	Flip            string `csv:"flip"`
	Linebreak2      string `csv:"linebreak2"`
	Players         string `csv:"players"`
	MoveInputs      string `csv:"move_inputs"`
	SpecialControls string `csv:"special_controls"`
	NumButtons      string `csv:"num_buttons"`
}

func readURL(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("download %s: unexpected HTTP status %s", url, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response from %s: %w", url, err)
	}
	return body, nil
}

func UpdateArcadeDB() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := &http.Client{Timeout: 30 * time.Second}

	body, err := readURL(ctx, client, config.ArcadeDBURL)
	if err != nil {
		return false, err
	}

	// A single file, so the contents API answers with one object rather than
	// the array the dated-folder layout used to return.
	var remote gitHubContentsItem
	if decodeErr := json.Unmarshal(body, &remote); decodeErr != nil {
		return false, fmt.Errorf("decode GitHub contents response: %w", decodeErr)
	}
	if remote.DownloadURL == "" || remote.SHA == "" {
		return false, nil
	}

	if mkdirErr := os.MkdirAll(config.MrextConfigFolder, 0o700); mkdirErr != nil {
		return false, fmt.Errorf("create metadata directory: %w", mkdirErr)
	}

	// The file name no longer carries a date, so freshness is decided by
	// comparing the local copy's Git blob hash with the one GitHub reports.
	// That is exact, needs nothing stored alongside the CSV, and a copy
	// truncated by a power cut or a full disk hashes differently and is
	// replaced rather than kept for good.
	local, hashErr := blobSHA(config.ArcadeDBFile)
	switch {
	case hashErr == nil && local == remote.SHA:
		return false, nil
	case hashErr != nil && !errors.Is(hashErr, fs.ErrNotExist):
		return false, hashErr
	}

	body, err = readURL(ctx, client, remote.DownloadURL)
	if err != nil {
		return false, err
	}
	if err := writeArcadeDB(body); err != nil {
		return false, err
	}

	return true, nil
}

// blobSHA computes a file's Git blob hash, which is what the GitHub contents
// API reports as a file's sha: sha1 over "blob <size>\x00" followed by the
// contents. Streamed, because this runs on a MiSTer and the database is
// hundreds of kilobytes.
func blobSHA(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat arcade database for hashing: %w", err)
	}
	// #nosec G304 -- fixed path from config, not user input.
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open arcade database for hashing: %w", err)
	}
	defer func() { _ = file.Close() }()

	// #nosec G401 -- Git object hashing is defined as SHA-1; not security use.
	hash := sha1.New()
	if _, err := fmt.Fprintf(hash, "blob %d\x00", info.Size()); err != nil {
		return "", fmt.Errorf("hash arcade database header: %w", err)
	}
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("hash arcade database contents: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// writeArcadeDB stages the download beside the destination and renames it
// over, so losing power mid-write leaves the previous copy intact.
func writeArcadeDB(body []byte) error {
	path := config.ArcadeDBFile
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create staged arcade database: %w", err)
	}
	staged := temporary.Name()
	published := false
	defer func() {
		_ = temporary.Close()
		if !published {
			_ = os.Remove(staged)
		}
	}()

	if _, err := temporary.Write(body); err != nil {
		return fmt.Errorf("write staged arcade database: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync staged arcade database: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close staged arcade database: %w", err)
	}
	// #nosec G703 -- destination is the fixed arcade database path.
	if err := os.Rename(staged, path); err != nil {
		return fmt.Errorf("publish arcade database: %w", err)
	}
	published = true
	return nil
}

func ReadArcadeDB() ([]ArcadeDBEntry, error) {
	if _, err := os.Stat(config.ArcadeDBFile); err != nil {
		return nil, fmt.Errorf("stat arcade database: %w", err)
	}

	dbFile, err := os.Open(config.ArcadeDBFile)
	if err != nil {
		return nil, fmt.Errorf("open arcade database: %w", err)
	}
	defer func() { _ = dbFile.Close() }()

	entries := make([]ArcadeDBEntry, 0)
	if err := gocsv.Unmarshal(dbFile, &entries); err != nil {
		return nil, fmt.Errorf("decode arcade database: %w", err)
	}
	return entries, nil
}
