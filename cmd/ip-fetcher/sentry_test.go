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

func SentryCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "sentry"})
}

func TestSentryCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		SentryCmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestSentryCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestSentryCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()
	t.Setenv("IP_FETCHER_MOCK_SENTRY", "true")

	app := mainpkg.GetApp()

	os.Args = []string{"ip-fetcher", "sentry", "--Path", filepath.Join(tDir, "test.txt")}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "test.txt"))

	os.Args = []string{"ip-fetcher", "sentry", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "sentry.txt"))
}

func TestSentryCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_SENTRY", "true")

	out := captureStdout(t, []string{"ip-fetcher", "sentry", "--stdout"})
	require.Contains(t, out, "34.123.33.225")
}

func TestSentryCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_SENTRY", "true")

	tDir := t.TempDir()

	out := captureStdout(t, []string{"ip-fetcher", "sentry", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "34.123.33.225")
	require.FileExists(t, filepath.Join(tDir, "sentry-prefixes.txt"))
}
