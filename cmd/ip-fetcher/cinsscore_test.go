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

func CinsscoreCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "cinsscore"})
}

func TestCinsscoreCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		CinsscoreCmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCinsscoreCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestCinsscoreCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.txt"
	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_CINSSCORE", "true")

	app := mainpkg.GetApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "cinsscore", "--Path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "cinsscore", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "cinsscore.txt"))
}

func TestCinsscoreCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_CINSSCORE", "true")

	out := captureStdout(t, []string{"ip-fetcher", "cinsscore", "--stdout"})
	require.Contains(t, out, "1.119.158.77")
}

// --lines renders the parsed document rather than the raw feed.
func TestCinsscoreCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_CINSSCORE", "true")

	tDir := t.TempDir()

	out := captureStdout(t,
		[]string{"ip-fetcher", "cinsscore", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "1.119.158.77")
	require.FileExists(t, filepath.Join(tDir, "cinsscore-prefixes.txt"))
}
