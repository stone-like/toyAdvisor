package main

import (
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/karrick/godirwalk"
	"github.com/pkg/errors"
)

func FormatFloat(val float64, precision int) string {
	return strconv.FormatFloat(val, 'f', precision, 64)
}

func GetInterval(cur, prev time.Time) int64 {
	return cur.Sub(prev).Nanoseconds()
}

func FileExists(file string) bool {
	if _, err := os.Stat(file); err != nil {
		return false
	}
	return true
}

func readString(dirpath string, file string) string {
	path := filepath.Join(dirpath, file)
	out, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

func readUInt64(dirpath string, file string) uint64 {

	out := readString(dirpath, file)
	if out == "" {
		return 0
	}

	val, err := strconv.ParseUint(out, 10, 64)
	if err != nil {
		return 0
	}

	return val
}

func ListDirectories(dirpath string, parent string, recursive bool, output map[string]struct{}) error {
	buf := make([]byte, godirwalk.MinimumScratchBufferSize)
	return listDirectories(dirpath, parent, recursive, output, buf)
}

func listDirectories(dirpath string, parent string, recursive bool, output map[string]struct{}, buf []byte) error {
	dirents, err := godirwalk.ReadDirents(dirpath, buf)
	if err != nil {
		// Ignore if this hierarchy does not exist.
		if os.IsNotExist(errors.Cause(err)) {
			err = nil
		}
		return err
	}
	for _, dirent := range dirents {
		// We only grab directories.
		if !dirent.IsDir() {
			continue
		}
		dirname := dirent.Name()

		name := path.Join(parent, dirname)
		output[name] = struct{}{}

		// List subcontainers if asked to.
		if recursive {
			err := listDirectories(path.Join(dirpath, dirname), name, true, output, buf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
