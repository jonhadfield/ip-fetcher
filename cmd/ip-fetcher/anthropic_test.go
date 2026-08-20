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

func AnthropicCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "anthropic"})
}

func TestAnthropicCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		AnthropicCmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestAnthropicCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestAnthropicCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.json"

	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_ANTHROPIC", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_ANTHROPIC")

	app := mainpkg.GetApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "anthropic", "--Path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "anthropic", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "anthropic.json"))
}

func TestAnthropicCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_ANTHROPIC", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_ANTHROPIC")

	out := captureStdout(t, []string{"ip-fetcher", "anthropic", "--stdout"})
	require.Contains(t, out, "216.73.216.0/22")
	require.Contains(t, out, "creationTime")
}

// --lines exercises docToLines against a Doc holding a time.Time.
func TestAnthropicCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_ANTHROPIC", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_ANTHROPIC")

	tDir := t.TempDir()

	out := captureStdout(t,
		[]string{"ip-fetcher", "anthropic", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "216.73.216.0/22")
	require.NotContains(t, out, "creationTime")
	require.FileExists(t, filepath.Join(tDir, "anthropic-prefixes.txt"))
}
