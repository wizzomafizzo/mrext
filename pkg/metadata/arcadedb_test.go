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

//nolint:gosec // Tests only operate on temporary fixture paths.
package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestUpdateArcadeDBUsesBlobHashAndReplacesTruncatedCopy(t *testing.T) {
	contents := []byte("setname,name\npooyan,Pooyan\n")
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.csv")
	if err := os.WriteFile(sourcePath, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	sha, err := blobSHA(sourcePath)
	if err != nil {
		t.Fatal(err)
	}

	var downloads atomic.Int32
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/contents":
			response := gitHubContentsItem{
				SHA:         sha,
				DownloadURL: server.URL + "/download",
			}
			if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
				t.Errorf("encode contents response: %v", encodeErr)
			}
		case "/download":
			downloads.Add(1)
			_, _ = w.Write(contents)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	databaseDir := filepath.Join(dir, "config")
	databasePath := filepath.Join(databaseDir, "ArcadeDatabase.csv")
	updated, err := updateArcadeDBFromSource(
		context.Background(),
		server.Client(),
		server.URL+"/contents",
		databaseDir,
		databasePath,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("missing database was not downloaded")
	}
	assertFileContents(t, databasePath, contents)

	updated, err = updateArcadeDBFromSource(
		context.Background(),
		server.Client(),
		server.URL+"/contents",
		databaseDir,
		databasePath,
	)
	if err != nil {
		t.Fatal(err)
	}
	if updated {
		t.Fatal("matching database was downloaded again")
	}
	if got := downloads.Load(); got != 1 {
		t.Fatalf("download requests = %d, want 1", got)
	}

	if writeErr := os.WriteFile(databasePath, contents[:8], 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	updated, err = updateArcadeDBFromSource(
		context.Background(),
		server.Client(),
		server.URL+"/contents",
		databaseDir,
		databasePath,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("truncated database was not replaced")
	}
	assertFileContents(t, databasePath, contents)
	if got := downloads.Load(); got != 2 {
		t.Fatalf("download requests = %d, want 2", got)
	}
}

func assertFileContents(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("file contents = %q, want %q", got, want)
	}
}
