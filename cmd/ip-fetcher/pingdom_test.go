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

func PingdomCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "pingdom"})
}

func TestPingdomCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		PingdomCmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestPingdomCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestPingdomCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()
	t.Setenv("IP_FETCHER_MOCK_PINGDOM", "true")

	app := mainpkg.GetApp()

	os.Args = []string{"ip-fetcher", "pingdom", "--Path", filepath.Join(tDir, "test.json")}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "test.json"))

	os.Args = []string{"ip-fetcher", "pingdom", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "pingdom.json"))
}

func TestPingdomCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_PINGDOM", "true")

	out := captureStdout(t, []string{"ip-fetcher", "pingdom", "--stdout"})
	require.Contains(t, out, "3.10.222.182")
}

func TestPingdomCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_PINGDOM", "true")

	tDir := t.TempDir()

	out := captureStdout(t, []string{"ip-fetcher", "pingdom", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "3.10.222.182")
	require.FileExists(t, filepath.Join(tDir, "pingdom-prefixes.txt"))
}
