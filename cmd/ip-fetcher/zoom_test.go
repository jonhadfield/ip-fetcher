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

func ZoomCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "zoom"})
}

func TestZoomCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		ZoomCmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestZoomCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestZoomCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()
	t.Setenv("IP_FETCHER_MOCK_ZOOM", "true")

	app := mainpkg.GetApp()

	os.Args = []string{"ip-fetcher", "zoom", "--Path", filepath.Join(tDir, "test.txt")}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "test.txt"))

	os.Args = []string{"ip-fetcher", "zoom", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "zoom.txt"))
}

func TestZoomCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_ZOOM", "true")

	out := captureStdout(t, []string{"ip-fetcher", "zoom", "--stdout"})
	require.Contains(t, out, "3.7.35.0/25")
}

func TestZoomCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_ZOOM", "true")

	tDir := t.TempDir()

	out := captureStdout(t, []string{"ip-fetcher", "zoom", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "3.7.35.0/25")
	require.FileExists(t, filepath.Join(tDir, "zoom-prefixes.txt"))
}
