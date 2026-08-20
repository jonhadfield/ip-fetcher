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

func DshieldCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "dshield"})
}

func TestDshieldCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		DshieldCmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestDshieldCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestDshieldCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.txt"
	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_DSHIELD", "true")

	app := mainpkg.GetApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "dshield", "--Path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "dshield", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "dshield.txt"))
}

func TestDshieldCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_DSHIELD", "true")

	out := captureStdout(t, []string{"ip-fetcher", "dshield", "--stdout"})
	require.Contains(t, out, "205.210.31.0")
}

// --lines renders the parsed document rather than the raw feed.
func TestDshieldCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_DSHIELD", "true")

	tDir := t.TempDir()

	out := captureStdout(t,
		[]string{"ip-fetcher", "dshield", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "205.210.31.0")
	require.FileExists(t, filepath.Join(tDir, "dshield-prefixes.txt"))
}
