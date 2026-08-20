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

func SpamhausCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "spamhaus"})
}

func TestSpamhausCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		SpamhausCmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestSpamhausCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestSpamhausCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.json"

	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_SPAMHAUS", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_SPAMHAUS")

	app := mainpkg.GetApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "spamhaus", "--Path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "spamhaus", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "spamhaus.json"))
}

func TestSpamhausCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_SPAMHAUS", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_SPAMHAUS")

	out := captureStdout(t, []string{"ip-fetcher", "spamhaus", "--stdout"})
	require.Contains(t, out, "1.10.16.0/20")
	require.Contains(t, out, "SBL256894")
}

// --lines emits bare prefixes for both address families, dropping the sblid and
// rir metadata.
func TestSpamhausCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_SPAMHAUS", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_SPAMHAUS")

	tDir := t.TempDir()

	out := captureStdout(t,
		[]string{"ip-fetcher", "spamhaus", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "1.10.16.0/20")
	require.Contains(t, out, "2001:678:254::/48")
	require.NotContains(t, out, "SBL256894")
	require.FileExists(t, filepath.Join(tDir, "spamhaus-prefixes.txt"))
}
