package main

import (
	"os"
	"path/filepath"
)

func saveFile(i saveFileInput) (path string, err error) {
	pf, err := os.Stat(i.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return path, err
		}
	}

	if pf != nil && pf.IsDir() {
		i.path = filepath.Join(i.path, i.defaultFileName)
	}

	f, err := os.Create(i.path)
	if err != nil {
		return path, err
	}

	defer f.Close()

	if _, err = f.Write(i.data); err != nil {
		return path, err
	}

	return filepath.Abs(i.path)
}

type saveFileInput struct {
	provider        string
	data            []byte
	path            string
	defaultFileName string
}
