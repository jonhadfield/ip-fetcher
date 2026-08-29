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

func StatuscakeCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "statuscake"})
}

func TestStatuscakeCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		StatuscakeCmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestStatuscakeCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestStatuscakeCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()
	t.Setenv("IP_FETCHER_MOCK_STATUSCAKE", "true")

	app := mainpkg.GetApp()

	os.Args = []string{"ip-fetcher", "statuscake", "--Path", filepath.Join(tDir, "test.json")}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "test.json"))

	os.Args = []string{"ip-fetcher", "statuscake", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "statuscake.json"))
}

func TestStatuscakeCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_STATUSCAKE", "true")

	out := captureStdout(t, []string{"ip-fetcher", "statuscake", "--stdout"})
	require.Contains(t, out, "146.190.20.113")
}

func TestStatuscakeCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_STATUSCAKE", "true")

	tDir := t.TempDir()

	out := captureStdout(t, []string{"ip-fetcher", "statuscake", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "146.190.20.113")
	require.FileExists(t, filepath.Join(tDir, "statuscake-prefixes.txt"))
}
