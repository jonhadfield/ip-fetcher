package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileSave(t *testing.T) {
	tDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tDir, "f1", "f2"), 0o755))

	// save file with dir + filename
	_, err := saveFile(saveFileInput{
		provider:        "example-provider",
		data:            []byte("some example data"),
		path:            filepath.Join(tDir, "file.json"),
		defaultFileName: "default.json",
	})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(tDir, "file.json"))

	// save file with dir only
	_, err = saveFile(saveFileInput{
		provider:        "example-provider",
		data:            []byte("some example data"),
		path:            tDir,
		defaultFileName: "default.json",
	})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(tDir, "default.json"))

	// save file in non-existant directory
	_, err = saveFile(saveFileInput{
		provider:        "example-provider",
		data:            []byte("some example data"),
		path:            filepath.Join(tDir, "bad-dir", "file.json"),
		defaultFileName: "default.json",
	})
	require.Error(t, err)

	// save file with just path separator
	_, err = saveFile(saveFileInput{
		provider:        "example-provider",
		data:            []byte("some example data"),
		path:            filepath.Join(tDir, string(os.PathSeparator)),
		defaultFileName: "default.json",
	})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(tDir, "default.json"))
}
