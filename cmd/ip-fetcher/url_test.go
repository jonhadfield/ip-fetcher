package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/agiledragon/gomonkey/v2"
	_ "github.com/agiledragon/gomonkey/v2/test/fake"
	"github.com/stretchr/testify/require"
)

const TestUrlAddr = "https://www.example.com/files/ips.txt"

func UrlCmdNoStdOutNoPath() {
	app := getApp()
	_ = app.Run([]string{"ip-fetcher", "url"})
}

func TestUrlCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		UrlCmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestUrlCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func UrlCmdEmptyPath() {
	defer testCleanUp(os.Args)

	app := getApp()
	_ = app.Run([]string{"ip-fetcher", "url"})
}

func TestUrlCmdEmptyPath(t *testing.T) {
	defer testCleanUp(os.Args)

	defer os.Unsetenv("TEST_EXIT")
	if os.Getenv("TEST_EXIT") == "1" {
		UrlCmdEmptyPath()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestUrlCmdEmptyPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestUrlCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "ips.txt"

	tDir := t.TempDir()

	_ = os.Setenv("IP_FETCHER_MOCK_URL", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_URL")

	app := getApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "url", "--path", filepath.Join(tDir, testFile), TestUrlAddr}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "url", "--path", tDir, TestUrlAddr}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "ips.txt"))
}

func TestUrlCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	_ = os.Setenv("IP_FETCHER_MOCK_URL", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_URL")

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

	app := getApp()
	os.Args = []string{"ip-fetcher", "url", "--stdout", TestUrlAddr}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "9.9.9.0/24")
}

func TestUrlCmdStdOutAndFile(t *testing.T) {
	defer testCleanUp(os.Args)
	_ = os.Setenv("IP_FETCHER_MOCK_URL", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_URL")

	tDir := t.TempDir()

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

	app := getApp()
	os.Args = []string{"ip-fetcher", "url", "--stdout", "--path", tDir, TestUrlAddr}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "9.9.9.0/24")
	require.FileExists(t, filepath.Join(tDir, "ips.txt"))
}
