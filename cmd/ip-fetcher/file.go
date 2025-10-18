package main

import (
	"os"
	"path/filepath"
)

func SaveFile(i SaveFileInput) (string, error) {
	if fi, err := os.Stat(i.Path); err == nil {
		if fi.IsDir() {
			i.Path = filepath.Join(i.Path, i.DefaultFileName)
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}

	if err := os.WriteFile(i.Path, i.Data, 0o600); err != nil {
		return "", err
	}

	return filepath.Abs(i.Path)
}

type SaveFileInput struct {
	Provider        string
	Data            []byte
	Path            string
	DefaultFileName string
}
