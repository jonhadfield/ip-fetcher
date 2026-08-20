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

func UptimerobotCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "uptimerobot"})
}

func TestUptimerobotCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		UptimerobotCmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestUptimerobotCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestUptimerobotCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.txt"

	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_UPTIMEROBOT", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_UPTIMEROBOT")

	app := mainpkg.GetApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "uptimerobot", "--Path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "uptimerobot", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "uptimerobot.txt"))
}

// the upstream list holds bare addresses; --lines normalises them to prefixes.
func TestUptimerobotCmdStdOutAndLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_UPTIMEROBOT", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_UPTIMEROBOT")

	raw := captureStdout(t, []string{"ip-fetcher", "uptimerobot", "--stdout"})
	require.Contains(t, raw, "3.12.251.153")
	require.NotContains(t, raw, "3.12.251.153/32")

	tDir := t.TempDir()

	lines := captureStdout(t,
		[]string{"ip-fetcher", "uptimerobot", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, lines, "3.12.251.153/32")
	require.Contains(t, lines, "2607:ff68:107::33/128")
	require.FileExists(t, filepath.Join(tDir, "uptimerobot-prefixes.txt"))
}
