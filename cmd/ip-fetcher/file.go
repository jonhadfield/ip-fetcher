package main

import (
	"os"
	"path/filepath"
)

func SaveFile(i SaveFileInput) (string, error) {
	pf, err := os.Stat(i.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
	}

	if pf != nil && pf.IsDir() {
		i.Path = filepath.Join(i.Path, i.DefaultFileName)
	}

	f, err := os.Create(i.Path)
	if err != nil {
		return "", err
	}

	defer f.Close()

	if _, err = f.Write(i.Data); err != nil {
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
