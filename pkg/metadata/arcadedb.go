package metadata

import (
	"encoding/json"
	"github.com/gocarina/gocsv"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"io"
	"net/http"
	"os"
	"time"
)

type GithubContentsItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	Url         string `json:"url"`
	HtmlUrl     string `json:"html_url"`
	GitUrl      string `json:"git_url"`
	DownloadUrl string `json:"download_url"`
	Type        string `json:"type"`
	Links       struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		Html string `json:"html"`
	} `json:"_links"`
}

type ArcadeDbEntry struct {
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

func UpdateArcadeDb() (bool, error) {
	resp, err := http.Get(config.ArcadeDBUrl)
	if err != nil {
		return false, err
	}
	defer func(b io.ReadCloser) {
		_ = b.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var contents []GithubContentsItem
	err = json.Unmarshal(body, &contents)
	if err != nil {
		return false, err
	} else if len(contents) == 0 {
		return false, nil
	}

	err = os.MkdirAll(config.MrextConfigFolder, 0755)
	if err != nil {
		return false, err
	}

	dbAge := time.Time{}
	if dbFile, err := os.Stat(config.ArcadeDBFile); err == nil {
		dbAge = dbFile.ModTime()
	}

	latestFile := contents[len(contents)-1]

	latestFileDate, err := time.Parse("ArcadeDatabase060102.csv", latestFile.Name)
	if err != nil {
		return false, err
	}

	if latestFileDate.Before(dbAge) {
		return false, nil
	}

	resp, err = http.Get(latestFile.DownloadUrl)
	if err != nil {
		return false, err
	}
	defer func(b io.ReadCloser) {
		_ = b.Close()
	}(resp.Body)

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	err = os.WriteFile(config.ArcadeDBFile, body, 0644)
	if err != nil {
		return false, err
	}

	return true, nil
}

func ReadArcadeDb() ([]ArcadeDbEntry, error) {
	if _, err := os.Stat(config.ArcadeDBFile); os.IsNotExist(err) {
		return nil, err
	}

	dbFile, err := os.Open(config.ArcadeDBFile)
	if err != nil {
		return nil, err
	}
	defer func(c io.Closer) {
		_ = c.Close()
	}(dbFile)

	entries := make([]ArcadeDbEntry, 0)
	err = gocsv.Unmarshal(dbFile, &entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}
