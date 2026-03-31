package helper

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
)

func SaveToFile(dir string, filename string, data []byte) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	fullPath := filepath.Join(dir, filename)

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return err
	}

	return nil
}

func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	os.RemoveAll(dest)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		in, err := f.Open()
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.Create(fpath)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err := io.Copy(out, in); err != nil {
			return err
		}
	}

	return nil
}

func FindProjectRoot(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return path, err
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		} else {
			return path, nil
		}
	}

	if len(dirs) == 1 {
		return filepath.Join(path, dirs[0]), nil
	}

	return path, nil
}

func FormatSizeRounded(paths []string) int {
	const KB int64 = 1024

	var size int64
	for _, path := range paths {
		size += getFileSize(path)
	}

	return int(size / KB)
}

func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func IsPRDirExist(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
