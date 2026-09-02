package main_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/agiledragon/gomonkey/v2"
	_ "github.com/agiledragon/gomonkey/v2/test/fake"
	mainpkg "github.com/jonhadfield/ip-fetcher/cmd/ip-fetcher"
	"github.com/stretchr/testify/require"
)

func UptrendsCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "uptrends"})
}

func TestUptrendsCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		UptrendsCmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestUptrendsCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestUptrendsCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()
	t.Setenv("IP_FETCHER_MOCK_UPTRENDS", "true")

	app := mainpkg.GetApp()

	os.Args = []string{"ip-fetcher", "uptrends", "--Path", filepath.Join(tDir, "test.json")}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "test.json"))

	os.Args = []string{"ip-fetcher", "uptrends", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "uptrends.json"))
}

func TestUptrendsCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_UPTRENDS", "true")

	out := captureStdout(t, []string{"ip-fetcher", "uptrends", "--stdout"})
	require.Contains(t, out, "101.201.208.194")
}

func TestUptrendsCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_UPTRENDS", "true")

	tDir := t.TempDir()

	out := captureStdout(t, []string{"ip-fetcher", "uptrends", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "101.201.208.194")
	require.FileExists(t, filepath.Join(tDir, "uptrends-prefixes.txt"))
}
