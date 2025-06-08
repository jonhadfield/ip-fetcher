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
	"github.com/jonhadfield/ip-fetcher/providers/digitalocean"
	"github.com/stretchr/testify/require"
)

func DigitaloceanCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "digitalocean"})
}

func TestDigitaloceanCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	defer os.Unsetenv("TEST_EXIT")
	if os.Getenv("TEST_EXIT") == "1" {
		DigitaloceanCmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestDigitaloceanCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestDigitaloceanCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.csv"

	tDir := t.TempDir()
	t.Setenv("IP_FETCHER_MOCK_DIGITALOCEAN", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_DIGITALOCEAN")

	ac := digitalocean.New()
	ac.DownloadURL = digitalocean.DigitaloceanDownloadURL

	app := mainpkg.GetApp()

	// with directory and filename
	os.Args = []string{"ip-fetcher", "digitalocean", "--Path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "digitalocean", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "google.csv"))
}

func TestDigitaloceanCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)

	t.Setenv("IP_FETCHER_MOCK_DIGITALOCEAN", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_DIGITALOCEAN")

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
	os.Args = []string{"ip-fetcher", "digitalocean", "--stdout"}

	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC

	require.Contains(t, out, "207.154.192.0/20,DE,DE-HE,Frankfurt,60342")
}

func TestDigitaloceanCmdStdOutAndFile(t *testing.T) {
	defer testCleanUp(os.Args)

	t.Setenv("IP_FETCHER_MOCK_DIGITALOCEAN", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_DIGITALOCEAN")

	tDir := t.TempDir()

	// stdout and file
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
	os.Args = []string{"ip-fetcher", "digitalocean", "--stdout", "--Path", tDir}

	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC

	require.FileExists(t, filepath.Join(tDir, "google.csv"))
	require.Contains(t, out, "207.154.192.0/20,DE,DE-HE,Frankfurt,60342")
}
