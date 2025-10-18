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

const (
	IPsV4Filename = "ips-v4"
	IPsV6Filename = "ips-v6"
)

func CloudflareCmdNoStdOutNoPaths() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "cloudflare"})
}

func TestCloudflareCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		CloudflareCmdNoStdOutNoPaths()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCloudflareCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")

	err := cmd.Run()
	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestCloudflareCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	t.Setenv("IP_FETCHER_MOCK_CLOUDFLARE", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_CLOUDFLARE")

	tDir := t.TempDir()

	app := mainpkg.GetApp()

	// with 4 only
	os.Args = []string{"ip-fetcher", "cloudflare", "--Path", tDir, "-4"}
	require.NoError(t, app.Run(os.Args))
	require.NoFileExists(t, filepath.Join(tDir, IPsV6Filename))
	require.FileExists(t, filepath.Join(tDir, IPsV4Filename))

	// new tmpDir now that previous test polluted it
	tDir = t.TempDir()

	// with 6 only
	os.Args = []string{"ip-fetcher", "cloudflare", "--Path", tDir, "-6"}
	require.NoError(t, app.Run(os.Args))
	require.NoFileExists(t, filepath.Join(tDir, IPsV4Filename))
	require.FileExists(t, filepath.Join(tDir, IPsV6Filename))

	// new tmpDir now that previous test polluted it
	tDir = t.TempDir()

	// with 4 & 6
	os.Args = []string{"ip-fetcher", "cloudflare", "--Path", tDir, "-4", "-6"}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, IPsV4Filename))
	require.FileExists(t, filepath.Join(tDir, IPsV6Filename))
}

func TestCloudflareCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_CLOUDFLARE", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_CLOUDFLARE")

	// stdout only
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
	os.Args = []string{"ip-fetcher", "cloudflare", "--stdout"}

	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC

	require.Contains(t, out, "131.0.72.1/22")
	require.Contains(t, out, "2a06:98c0::/29")
	require.NoFileExists(t, tDir, "ips-v4")
	require.NoFileExists(t, tDir, "ips-v6")
}

func TestCloudflareCmdStdOutAndFiles(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_CLOUDFLARE", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_CLOUDFLARE")

	// stdout only
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
	os.Args = []string{"ip-fetcher", "cloudflare", "--stdout", "--Path", tDir}

	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC

	require.Contains(t, out, "131.0.72.1/22")
	require.Contains(t, out, "2a06:98c0::/29")

	require.FileExists(t, filepath.Join(tDir, IPsV4Filename))
	require.FileExists(t, filepath.Join(tDir, IPsV6Filename))
}
