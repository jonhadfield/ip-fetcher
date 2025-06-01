package main_test

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

func GoogleCmdNoStdOutNoPath() {
	app := getApp()
	_ = app.Run([]string{"ip-fetcher", "google"})
}

func TestGoogleCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		GoogleCmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGoogleCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func GoogleCmdEmptyPath() {
	defer testCleanUp(os.Args)

	app := getApp()
	_ = app.Run([]string{"ip-fetcher", "google"})
}

func TestGoogleCmdEmptyPath(t *testing.T) {
	defer testCleanUp(os.Args)

	defer os.Unsetenv("TEST_EXIT")
	if os.Getenv("TEST_EXIT") == "1" {
		GoogleCmdEmptyPath()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestGoogleCmdEmptyPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestGoogleCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.json"

	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_GOOGLE", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_GOOGLE")

	app := getApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "google", "--path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "google", "--path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "goog.json"))
}

func TestGoogleCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_GOOGLE", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_GOOGLE")

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
	os.Args = []string{"ip-fetcher", "google", "--stdout"}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "2620:120:e001::/40")
}

func TestGoogleCmdStdOutAndFile(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_GOOGLE", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_GOOGLE")

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
	os.Args = []string{"ip-fetcher", "google", "--stdout", "--path", tDir}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "34.3.4.0/24")
	require.FileExists(t, filepath.Join(tDir, "goog.json"))
}
