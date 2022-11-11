package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

type MusicState struct {
	Playing  bool   `json:"playing"`
	Playback string `json:"playback"`
	Playlist string `json:"playlist"`
	Track    string `json:"track"`
}

type MusicServiceStatus struct {
	Running bool `json:"running"`
}

type MusicPlaylists []string

const musicFolder = config.SdFolder + "/music"
const musicSocket = "/tmp/bgm.sock"
const socketBuffer = 4096

func sendCmd(cmd string) (string, error) {
	conn, err := net.Dial("unix", musicSocket)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(cmd))
	if err != nil {
		return "", err
	}

	buf := make([]byte, socketBuffer)
	_, err = conn.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	return string(bytes.Trim(buf, "\x00")), nil
}

func musicStatus(w http.ResponseWriter, r *http.Request) {
	var state MusicState

	resp, err := sendCmd("status")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	states := strings.Split(resp, "\t")
	if len(states) < 4 {
		http.Error(w, "invalid response from bgm", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	state.Playing = states[0] == "yes"
	state.Playback = states[1]
	state.Playlist = states[2]
	state.Track = states[3]

	json.NewEncoder(w).Encode(state)
}

func musicPlay(w http.ResponseWriter, r *http.Request) {
	_, err := sendCmd("play")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func musicStop(w http.ResponseWriter, r *http.Request) {
	_, err := sendCmd("stop")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func musicSkip(w http.ResponseWriter, r *http.Request) {
	_, err := sendCmd("skip")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func setMusicPlayback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playback := vars["playback"]

	_, err := sendCmd("set playback " + playback)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func setMusicPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playlist := vars["playlist"]

	_, err := sendCmd("set playlist " + playlist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func musicPlaylists(w http.ResponseWriter, r *http.Request) {
	var playlists MusicPlaylists

	items, err := os.ReadDir(musicFolder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	for _, item := range items {
		if item.IsDir() && item.Name() != "boot" {
			playlists = append(playlists, item.Name())
		}
	}

	json.NewEncoder(w).Encode(playlists)
}

func musicServiceStatus(w http.ResponseWriter, r *http.Request) {
	var status MusicServiceStatus

	_, err := os.Stat(musicSocket)
	if err != nil {
		status.Running = false
	} else {
		status.Running = true
	}

	json.NewEncoder(w).Encode(status)
}
