package main_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/cloudflare"
	"github.com/stretchr/testify/require"
)

func CloudflareCmdNoStdOutNoPaths() {
	app := getApp()
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
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestCloudflareCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	t.Setenv("IP_FETCHER_MOCK_CLOUDFLARE", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_CLOUDFLARE")

	tDir := t.TempDir()

	ac := cloudflare.New()
	ac.IPv4DownloadURL = cloudflare.DefaultIPv4URL
	ac.IPv6DownloadURL = cloudflare.DefaultIPv6URL

	app := getApp()

	// with 4 only
	os.Args = []string{"ip-fetcher", "cloudflare", "--path", tDir, "-4"}
	require.NoError(t, app.Run(os.Args))
	require.NoFileExists(t, filepath.Join(tDir, ipsv6Filename))
	require.FileExists(t, filepath.Join(tDir, ipsv4Filename))

	// new tmpDir now that previous test polluted it
	tDir = t.TempDir()

	// with 6 only
	os.Args = []string{"ip-fetcher", "cloudflare", "--path", tDir, "-6"}
	require.NoError(t, app.Run(os.Args))
	require.NoFileExists(t, filepath.Join(tDir, ipsv4Filename))
	require.FileExists(t, filepath.Join(tDir, ipsv6Filename))

	// new tmpDir now that previous test polluted it
	tDir = t.TempDir()

	// with 4 & 6
	os.Args = []string{"ip-fetcher", "cloudflare", "--path", tDir, "-4", "-6"}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, ipsv4Filename))
	require.FileExists(t, filepath.Join(tDir, ipsv6Filename))
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

	app := getApp()
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

	ac := cloudflare.New()
	ac.IPv4DownloadURL = cloudflare.DefaultIPv4URL
	ac.IPv6DownloadURL = cloudflare.DefaultIPv6URL

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

	app := getApp()
	os.Args = []string{"ip-fetcher", "cloudflare", "--stdout", "--path", tDir}

	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC

	require.Contains(t, out, "131.0.72.1/22")
	require.Contains(t, out, "2a06:98c0::/29")

	require.FileExists(t, filepath.Join(tDir, ipsv4Filename))
	require.FileExists(t, filepath.Join(tDir, ipsv6Filename))
}
