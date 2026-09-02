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

func Site24x7CmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "site24x7"})
}

func TestSite24x7CmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		Site24x7CmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestSite24x7CmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestSite24x7CmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()
	t.Setenv("IP_FETCHER_MOCK_SITE24X7", "true")

	app := mainpkg.GetApp()

	os.Args = []string{"ip-fetcher", "site24x7", "--Path", filepath.Join(tDir, "test.json")}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "test.json"))

	os.Args = []string{"ip-fetcher", "site24x7", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "site24x7.json"))
}

func TestSite24x7CmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_SITE24X7", "true")

	out := captureStdout(t, []string{"ip-fetcher", "site24x7", "--stdout"})
	require.Contains(t, out, "37.221.111.107")
}

func TestSite24x7CmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_SITE24X7", "true")

	tDir := t.TempDir()

	out := captureStdout(t, []string{"ip-fetcher", "site24x7", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "37.221.111.107")
	require.FileExists(t, filepath.Join(tDir, "site24x7-prefixes.txt"))
}
