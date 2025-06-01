package main_test

import (
	"os"
	"path/filepath"
	"testing"

	mainpkg "github.com/jonhadfield/ip-fetcher/cmd/ip-fetcher"
	"github.com/stretchr/testify/require"
)

func TestFileSave(t *testing.T) {
	tDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tDir, "f1", "f2"), 0o755))

	// save file with dir + filename
	_, err := mainpkg.SaveFile(mainpkg.SaveFileInput{
		Provider:        "example-Provider",
		Data:            []byte("some example Data"),
		Path:            filepath.Join(tDir, "file.json"),
		DefaultFileName: "default.json",
	})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(tDir, "file.json"))

	// save file with dir only
	_, err = mainpkg.SaveFile(mainpkg.SaveFileInput{
		Provider:        "example-Provider",
		Data:            []byte("some example Data"),
		Path:            tDir,
		DefaultFileName: "default.json",
	})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(tDir, "default.json"))

	// save file in non-existant directory
	_, err = mainpkg.SaveFile(mainpkg.SaveFileInput{
		Provider:        "example-Provider",
		Data:            []byte("some example Data"),
		Path:            filepath.Join(tDir, "bad-dir", "file.json"),
		DefaultFileName: "default.json",
	})
	require.Error(t, err)

	// save file with just Path separator
	_, err = mainpkg.SaveFile(mainpkg.SaveFileInput{
		Provider:        "example-Provider",
		Data:            []byte("some example Data"),
		Path:            filepath.Join(tDir, string(os.PathSeparator)),
		DefaultFileName: "default.json",
	})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(tDir, "default.json"))
}
