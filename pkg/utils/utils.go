package utils

import (
	"archive/zip"
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
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

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func IsZip(path string) bool {
	// TODO: this should check the file header
	return filepath.Ext(strings.ToLower(path)) == ".zip"
}

// ListZip returns a slice of all filenames in a zip file.
func ListZip(path string) ([]string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var files []string
	for _, f := range r.File {
		files = append(files, f.Name)
	}

	return files, nil
}

func CopyFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return err
	}
	outputFile.Sync()
	inputFile.Close()

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
		return err
	}

	return nil
}

// Max returns the highest value in a slice.
func Max[T constraints.Ordered](xs []T) T {
	if len(xs) == 0 {
		var zv T
		return zv
	}
	max := xs[0]
	for _, x := range xs {
		if x > max {
			max = x
		}
	}
	return max
}

// Min returns the lowest value in a slice.
func Min[T constraints.Ordered](xs []T) T {
	if len(xs) == 0 {
		var zv T
		return zv
	}
	min := xs[0]
	for _, x := range xs {
		if x < min {
			min = x
		}
	}
	return min
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
		return item, fmt.Errorf("empty slice")
	} else {
		item = xs[r.Intn(len(xs))]
		return item, nil
	}
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

// SortedMapKeys return a sorted list of all keys in a map.
func SortedMapKeys[V any](m map[string]V) []string {
	keys := MapKeys(m)
	sort.Strings(keys)
	return keys
}

func StripChars(s string, chars string) string {
	for _, c := range chars {
		s = strings.ReplaceAll(s, string(c), "")
	}
	return s
}

// StripBadFileChars removes all characters from a string that are not allowed in filenames.
func StripBadFileChars(s string) string {
	return StripChars(s, "/\\:*?\"<>|")
}

// Md5Sum returns the MD5 hash of a file on disk.
func Md5Sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	io.Copy(hash, file)
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// YesOrNoPrompt displays a simple yes/no prompt for use with a controller.
func YesOrNoPrompt(prompt string) bool {
	fmt.Printf(prompt + " [DOWN=Yes/UP=No] ")

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, 3)
	reader.Read(buf)

	term.Restore(int(os.Stdin.Fd()), oldState)

	delay := func() { time.Sleep(400 * time.Millisecond) }

	if buf[0] == 27 && buf[1] == 91 && buf[2] == 66 {
		fmt.Println("Yes")
		delay()
		return true
	} else {
		// 27 91 65 is up arrow
		fmt.Println("No")
		delay()
		return false
	}
}

// InfoPrompt displays an information prompt for use with a controller.
func InfoPrompt(prompt string) {
	fmt.Println(prompt)
	fmt.Println("Press any key to continue...")

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, 1)
	reader.Read(buf)

	term.Restore(int(os.Stdin.Fd()), oldState)

	time.Sleep(400 * time.Millisecond)
}

func IsEmptyDir(path string) (bool, error) {
	dir, err := os.ReadDir(path)
	if err != nil {
		return false, err
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
		return err
	}

	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]

		empty, err := IsEmptyDir(dir)
		if err != nil {
			return err
		}

		if empty {
			err = os.Remove(dir)
			if err != nil {
				return err
			}
		}
	}

	rootEmpty, err := IsEmptyDir(path)
	if err != nil {
		return err
	} else if rootEmpty {
		err = os.Remove(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetLocalIp() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

func WaitForInternet(maxTries int) bool {
	for i := 0; i < maxTries; i++ {
		_, err := http.Get("https://api.github.com")
		if err == nil {
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func AlphaMapKeys[V any](m map[string]V) []string {
	keys := MapKeys(m)
	sort.Strings(keys)
	return keys
}

func Reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func RemoveFileExt(s string) string {
	parts := strings.Split(s, ".")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ".")
	}
	return s
}
