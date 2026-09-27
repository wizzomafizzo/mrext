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
package playlog

import (
	"sort"
	"time"
)

// Summary is the headline figures for a stats screen.
type Summary struct {
	FirstPlayed    *time.Time    `json:"firstPlayed"`
	LastPlayed     *time.Time    `json:"lastPlayed"`
	LongestSession *Session      `json:"longestSession"`
	TopGames       []GameStats   `json:"topGames"`
	TopSystems     []SystemStats `json:"topSystems"`
	Days           []DayStats    `json:"days"`
	Hours          []int         `json:"hours"`
	TotalTime      int           `json:"totalTime"`
	CoreTime       int           `json:"coreTime"`
	GamesPlayed    int           `json:"gamesPlayed"`
	Sessions       int           `json:"sessions"`
	CurrentStreak  int           `json:"currentStreak"`
	LongestStreak  int           `json:"longestStreak"`
}

// Summarise computes the summary. Days, hours and streaks are counted in
// zone, the client's, since the MiSTer usually runs on UTC; now ends the
// day range. top and days bound the lists.
func (h *History) Summarise(now time.Time, zone *time.Location, top, days int) Summary {
	summary := Summary{
		CoreTime:   h.CoreTime,
		Sessions:   len(h.Sessions),
		TopGames:   []GameStats{},
		TopSystems: []SystemStats{},
		Days:       []DayStats{},
		Hours:      make([]int, 24),
	}

	played := h.Played()
	systems := map[string]*SystemStats{}
	for _, game := range played {
		summary.TotalTime += game.Time
		if game.Time > 0 {
			summary.GamesPlayed++
		}
		entry, ok := systems[game.System]
		if !ok {
			entry = &SystemStats{System: game.System}
			systems[game.System] = entry
		}
		entry.Time += game.Time
		entry.Games++
	}

	SortByTime(played)
	summary.TopGames = played[:min(top, len(played))]

	for _, entry := range systems {
		summary.TopSystems = append(summary.TopSystems, *entry)
	}
	sort.Slice(summary.TopSystems, func(i, j int) bool {
		if summary.TopSystems[i].Time != summary.TopSystems[j].Time {
			return summary.TopSystems[i].Time > summary.TopSystems[j].Time
		}
		return summary.TopSystems[i].System < summary.TopSystems[j].System
	})
	summary.TopSystems = summary.TopSystems[:min(top, len(summary.TopSystems))]

	perDay := map[string]int{}
	for i := range h.Sessions {
		session := h.Sessions[i]
		start := session.Start.In(zone)
		perDay[start.Format(time.DateOnly)] += session.Duration
		summary.Hours[start.Hour()] += session.Duration
		if summary.FirstPlayed == nil || start.Before(*summary.FirstPlayed) {
			first := start
			summary.FirstPlayed = &first
		}
		if summary.LastPlayed == nil || start.After(*summary.LastPlayed) {
			last := start
			summary.LastPlayed = &last
		}
		if summary.LongestSession == nil || session.Duration > summary.LongestSession.Duration {
			longest := session
			summary.LongestSession = &longest
		}
	}

	today := now.In(zone)
	for offset := days - 1; offset >= 0; offset-- {
		date := today.AddDate(0, 0, -offset).Format(time.DateOnly)
		summary.Days = append(summary.Days, DayStats{Date: date, Time: perDay[date]})
	}
	summary.CurrentStreak, summary.LongestStreak = streaks(perDay, today)
	return summary
}

// streaks counts consecutive days with play. The current streak survives
// until a whole day passes without any, so it still counts this morning
// what was played last night.
func streaks(perDay map[string]int, today time.Time) (current, longest int) {
	dates := make([]string, 0, len(perDay))
	for date := range perDay {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	run := 0
	var previous time.Time
	for _, date := range dates {
		day, err := time.ParseInLocation(time.DateOnly, date, today.Location())
		if err != nil {
			continue
		}
		if run > 0 && day.Equal(previous.AddDate(0, 0, 1)) {
			run++
		} else {
			run = 1
		}
		previous = day
		longest = max(longest, run)
	}

	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	if _, playedToday := perDay[start.Format(time.DateOnly)]; !playedToday {
		start = start.AddDate(0, 0, -1)
	}
	for {
		if _, ok := perDay[start.Format(time.DateOnly)]; !ok {
			break
		}
		current++
		start = start.AddDate(0, 0, -1)
	}
	return current, longest
}
