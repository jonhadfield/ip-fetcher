package main_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	mainpkg "github.com/jonhadfield/ip-fetcher/cmd/ip-fetcher"
	"github.com/stretchr/testify/require"
)

func TestOpenAICmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.json"

	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_OPENAI", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_OPENAI")

	app := mainpkg.GetApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "openai", "--Path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "openai", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "openai.json"))
}

func TestOpenAICmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_OPENAI", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_OPENAI")

	// stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	app := mainpkg.GetApp()
	os.Args = []string{"ip-fetcher", "openai", "--stdout"}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "132.196.86.0/24")
	require.Contains(t, out, "104.210.140.128/28")
	require.Contains(t, out, "13.65.138.112/28")
	require.Contains(t, out, "2a01:111:f403:c111::/64")
}
