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
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/gamesdb"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

func searchResults(n int) []SearchResultGame {
	results := make([]SearchResultGame, 0, n)
	for i := range n {
		results = append(results, SearchResultGame{Name: fmt.Sprintf("game %04d", i)})
	}
	return results
}

func TestSearchPageWalksResultsInOrder(t *testing.T) {
	results := searchResults(2*pageSize + 201)

	pages := []struct {
		first string
		page  int
		count int
	}{
		{"game 0000", 1, pageSize},
		{"game 0500", 2, pageSize},
		{"game 1000", 3, 201},
		{"", 4, 0},
	}
	for _, want := range pages {
		got := searchPage(results, want.page)
		if len(got) != want.count {
			t.Fatalf("page %d: got %d results, want %d", want.page, len(got), want.count)
		}
		if want.count > 0 && got[0].Name != want.first {
			t.Fatalf("page %d: first result %q, want %q", want.page, got[0].Name, want.first)
		}
	}
}

func TestSearchPageEndsExactlyOnAFullPage(t *testing.T) {
	results := searchResults(pageSize)
	if got := len(searchPage(results, 1)); got != pageSize {
		t.Fatalf("page 1: got %d results, want %d", got, pageSize)
	}
	got := searchPage(results, 2)
	if got == nil {
		t.Fatal("page 2: got nil, want an empty list so it encodes as []")
	}
	if len(got) != 0 {
		t.Fatalf("page 2: got %d results, want 0", len(got))
	}
}

func TestSearchPageOfNoResultsIsEmpty(t *testing.T) {
	if got := searchPage(searchResults(0), 1); got == nil || len(got) != 0 {
		t.Fatalf("got %v, want an empty list", got)
	}
}

func TestSearchRejectsAPageBelowOne(t *testing.T) {
	handler := Search(service.NewLogger("mrext-search-page-test"))
	body := strings.NewReader(`{"query":"x","system":"NES","page":-1}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/games/search", body)
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSearchPageBeyondIntRangeIsEmpty(t *testing.T) {
	results := searchResults(pageSize + 1)
	for _, page := range []int{math.MaxInt, math.MaxInt / pageSize, 3} {
		if got := searchPage(results, page); len(got) != 0 {
			t.Fatalf("page %d: got %d results, want 0", page, len(got))
		}
	}
}

// The handler end to end: the index query is stood in with 1201 hits so the
// response carries the requested page, its slice of the results and the
// unchanged total.
func TestSearchHandlerReturnsTheRequestedPage(t *testing.T) {
	previous := searchNames
	t.Cleanup(func() { searchNames = previous })
	searchNames = func(systems []games.System, _ string) ([]gamesdb.SearchResult, error) {
		hits := make([]gamesdb.SearchResult, 0, 2*pageSize+201)
		for i := range 2*pageSize + 201 {
			hits = append(hits, gamesdb.SearchResult{
				SystemID: systems[0].Id,
				Name:     fmt.Sprintf("game %04d", i),
				Path:     fmt.Sprintf("/media/fat/games/NES/game %04d.nes", i),
			})
		}
		return hits, nil
	}
	handler := Search(service.NewLogger("mrext-search-page-test"))

	for _, want := range []struct {
		first string
		page  int
		count int
	}{
		{"game 0000", 1, pageSize},
		{"game 0500", 2, pageSize},
		{"game 1000", 3, 201},
		{"", 4, 0},
	} {
		body := strings.NewReader(fmt.Sprintf(`{"query":"game","system":"NES","page":%d}`, want.page))
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/games/search", body)
		rec := httptest.NewRecorder()
		handler(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("page %d: status %d, body %s", want.page, rec.Code, rec.Body.String())
		}
		var got SearchResults
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("page %d: decode: %v", want.page, err)
		}
		if got.Page != want.page || got.PageSize != pageSize || got.Total != 2*pageSize+201 {
			t.Fatalf("page %d: metadata page=%d pageSize=%d total=%d", want.page, got.Page, got.PageSize, got.Total)
		}
		if len(got.Data) != want.count {
			t.Fatalf("page %d: got %d results, want %d", want.page, len(got.Data), want.count)
		}
		if want.count > 0 && got.Data[0].Name != want.first {
			t.Fatalf("page %d: first result %q, want %q", want.page, got.Data[0].Name, want.first)
		}
	}

	// No page at all is page 1, which is what every older client sends.
	body := strings.NewReader(`{"query":"game","system":"NES"}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/games/search", body)
	rec := httptest.NewRecorder()
	handler(rec, req)
	var got SearchResults
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Page != 1 || len(got.Data) != pageSize {
		t.Fatalf("no page: page=%d results=%d", got.Page, len(got.Data))
	}
}
