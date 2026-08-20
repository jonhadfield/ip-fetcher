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

func AkamaiCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "akamai"})
}

func TestAkamaiCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		AkamaiCmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestAkamaiCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestAkamaiCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.txt"
	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_AKAMAI", "true")

	app := mainpkg.GetApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "akamai", "--Path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "akamai", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "prefixes.txt"))
}

// Akamai publishes a zip, so the command must write parsed prefixes rather
// than the raw response.
func TestAkamaiCmdStdOutWritesPrefixesNotArchive(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_AKAMAI", "true")

	out := captureStdout(t, []string{"ip-fetcher", "akamai", "--stdout"})

	require.Contains(t, out, "203.0.113.0/24")
	require.Contains(t, out, "198.51.100.0/24")
	require.Contains(t, out, "2001:db8::/32")

	// the zip's local file header and member names must not reach the output.
	require.NotContains(t, out, "PK\x03\x04")
	require.NotContains(t, out, "akamai_ipv4_CIDRs.txt")
}

// the saved file holds the same parsed prefixes as stdout.
func TestAkamaiCmdSavedFileContent(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_AKAMAI", "true")

	tDir := t.TempDir()

	app := mainpkg.GetApp()
	os.Args = []string{"ip-fetcher", "akamai", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))

	content, err := os.ReadFile(filepath.Join(tDir, "prefixes.txt"))
	require.NoError(t, err)
	require.Contains(t, string(content), "203.0.113.0/24")
	require.NotContains(t, string(content), "PK\x03\x04")
}
