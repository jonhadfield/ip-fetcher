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

func GreensnowCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "greensnow"})
}

func TestGreensnowCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		GreensnowCmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGreensnowCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestGreensnowCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_GREENSNOW", "true")

	app := mainpkg.GetApp()

	os.Args = []string{"ip-fetcher", "greensnow", "--Path", filepath.Join(tDir, "test.txt")}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "test.txt"))

	os.Args = []string{"ip-fetcher", "greensnow", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "greensnow.txt"))
}

func TestGreensnowCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_GREENSNOW", "true")

	out := captureStdout(t, []string{"ip-fetcher", "greensnow", "--stdout"})
	require.Contains(t, out, "79.124.56.146")
}

func TestGreensnowCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_GREENSNOW", "true")

	tDir := t.TempDir()

	out := captureStdout(t, []string{"ip-fetcher", "greensnow", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "79.124.56.146")
	require.FileExists(t, filepath.Join(tDir, "greensnow-prefixes.txt"))
}
