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

package utils

import (
	"archive/zip"
	"bufio"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/exp/constraints"
	"golang.org/x/term"
)

func IsZip(path string) bool {
	// TODO: this should check the file header
	return filepath.Ext(strings.ToLower(path)) == ".zip"
}

// ListZip returns a slice of all filenames in a zip file.
func ListZip(path string) ([]string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open ZIP file: %w", err)
	}

	files := make([]string, 0, len(r.File))
	for _, f := range r.File {
		files = append(files, f.Name)
	}

	if err := r.Close(); err != nil {
		return nil, fmt.Errorf("close ZIP file: %w", err)
	}
	return files, nil
}

func CopyFile(sourcePath, destPath string) error {
	// #nosec G304 -- both paths are explicit inputs to this filesystem utility.
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer func() { _ = inputFile.Close() }()

	// #nosec G304 -- both paths are explicit inputs to this filesystem utility.
	outputFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer func() { _ = outputFile.Close() }()

	if _, err := io.Copy(outputFile, inputFile); err != nil {
		return fmt.Errorf("copy file data: %w", err)
	}
	if err := outputFile.Sync(); err != nil {
		return fmt.Errorf("sync destination file: %w", err)
	}
	if err := outputFile.Close(); err != nil {
		return fmt.Errorf("close destination file: %w", err)
	}

	return nil
}

// MoveFile moves a file. Supports moving between filesystems.
func MoveFile(sourcePath, destPath string) error {
	err := CopyFile(sourcePath, destPath)
	if err != nil {
		return err
	}

	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("remove source file: %w", err)
	}

	return nil
}

// Max returns the highest value in a slice.
func Max[T constraints.Ordered](xs []T) T {
	if len(xs) == 0 {
		var zv T
		return zv
	}
	best := xs[0]
	for _, x := range xs {
		if x > best {
			best = x
		}
	}
	return best
}

// Min returns the lowest value in a slice.
func Min[T constraints.Ordered](xs []T) T {
	if len(xs) == 0 {
		var zv T
		return zv
	}
	best := xs[0]
	for _, x := range xs {
		if x < best {
			best = x
		}
	}
	return best
}

// Contains returns true if slice contains value.
func Contains[T comparable](xs []T, x T) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// ContainsFold returns true if slice of strings contains value (case insensitive).
func ContainsFold(xs []string, x string) bool {
	for _, v := range xs {
		if strings.EqualFold(v, x) {
			return true
		}
	}
	return false
}

// RandomElem picks and returns a random element from a slice.
func RandomElem[T any](xs []T) (T, error) {
	var item T
	if len(xs) == 0 {
		return item, errors.New("empty slice")
	}

	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(xs))))
	if err != nil {
		return item, fmt.Errorf("choose random element: %w", err)
	}
	return xs[index.Int64()], nil
}

// MapKeys returns a list of all keys in a map.
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

func StripChars(s, chars string) string {
	for _, c := range chars {
		s = strings.ReplaceAll(s, string(c), "")
	}
	return s
}

// StripBadFileChars removes all characters from a string that are not allowed in filenames.
func StripBadFileChars(s string) string {
	return StripChars(s, "/\\:*?\"<>|")
}

// YesOrNoPrompt displays a simple yes/no prompt for use with a controller.
func YesOrNoPrompt(prompt string) bool {
	_, _ = fmt.Printf("%s [DOWN=Yes/UP=No] ", prompt)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, 3)
	if _, err := io.ReadFull(reader, buf); err != nil {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
		return false
	}
	if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
		return false
	}

	delay := func() { time.Sleep(400 * time.Millisecond) }

	if buf[0] == 27 && buf[1] == 91 && buf[2] == 66 {
		_, _ = fmt.Println("Yes")
		delay()
		return true
	}

	// 27 91 65 is up arrow
	_, _ = fmt.Println("No")
	delay()
	return false
}

func IsEmptyDir(path string) (bool, error) {
	dir, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("read directory: %w", err)
	}

	return len(dir) == 0, nil
}

// RemoveEmptyDirs removes all empty folders in a path, including folders containing only empty
// folders and the path itself.
func RemoveEmptyDirs(path string) error {
	var dirs []string

	err := filepath.WalkDir(path, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			dirs = append(dirs, path)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walk directories: %w", err)
	}

	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]

		isEmpty, checkErr := IsEmptyDir(dir)
		if checkErr != nil {
			return checkErr
		}

		if isEmpty {
			if removeErr := os.Remove(dir); removeErr != nil {
				return fmt.Errorf("remove empty directory: %w", removeErr)
			}
		}
	}

	rootEmpty, err := IsEmptyDir(path)
	if err != nil {
		return err
	}
	if rootEmpty {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove empty root directory: %w", err)
		}
	}

	return nil
}

func GetLocalIP() (net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(ctx, "udp", "8.8.8.8:80")
	if err != nil {
		return nil, fmt.Errorf("discover local IP: %w", err)
	}
	defer func() { _ = conn.Close() }()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil, fmt.Errorf("unexpected local address type %T", conn.LocalAddr())
	}
	return localAddr.IP, nil
}

func WaitForInternet(maxTries int) bool {
	client := &http.Client{Timeout: 10 * time.Second}
	for range maxTries {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com", http.NoBody)
		if err != nil {
			cancel()
			return false
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			cancel()
			return true
		}
		cancel()
		time.Sleep(time.Second)
	}
	return false
}

func AlphaMapKeys[V any](m map[string]V) []string {
	keys := MapKeys(m)
	sort.Strings(keys)
	return keys
}

func RemoveFileExt(s string) string {
	parts := strings.Split(s, ".")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ".")
	}
	return s
}
