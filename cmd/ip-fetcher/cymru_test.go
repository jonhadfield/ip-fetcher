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

func CymruCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "cymru"})
}

func TestCymruCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		CymruCmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCymruCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestCymruCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_CYMRU", "true")

	app := mainpkg.GetApp()

	os.Args = []string{"ip-fetcher", "cymru", "--Path", filepath.Join(tDir, "test.json")}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "test.json"))

	os.Args = []string{"ip-fetcher", "cymru", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "cymru.json"))
}

func TestCymruCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_CYMRU", "true")

	out := captureStdout(t, []string{"ip-fetcher", "cymru", "--stdout"})
	require.Contains(t, out, "0.0.0.0/8")
}

func TestCymruCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_CYMRU", "true")

	tDir := t.TempDir()

	out := captureStdout(t, []string{"ip-fetcher", "cymru", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "0.0.0.0/8")
	require.FileExists(t, filepath.Join(tDir, "cymru-prefixes.txt"))
}
