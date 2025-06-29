package main_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	mainpkg "github.com/jonhadfield/ip-fetcher/cmd/ip-fetcher"
	"github.com/stretchr/testify/require"
)

func OVHCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "ovh"})
}

func TestOVHCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	defer os.Unsetenv("TEST_EXIT")
	if os.Getenv("TEST_EXIT") == "1" {
		OVHCmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestOVHCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestOVHCmdSaveToPath(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()
	t.Setenv("IP_FETCHER_MOCK_OVH", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_OVH")

	app := mainpkg.GetApp()

	os.Args = []string{"ip-fetcher", "ovh", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "prefixes.txt"))
}

func TestOVHCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)

	t.Setenv("IP_FETCHER_MOCK_OVH", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_OVH")

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
	os.Args = []string{"ip-fetcher", "ovh", "--stdout"}

	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC

	require.Contains(t, out, "192.0.2.0/24")
}
