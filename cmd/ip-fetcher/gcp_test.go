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

func GCPCmdNoStdOutNoPath() {
	app := getApp()
	_ = app.Run([]string{"ip-fetcher", "gcp"})
}

func TestGCPCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		GCPCmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGCPCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func GCPCmdEmptyPath() {
	defer testCleanUp(os.Args)

	app := getApp()
	_ = app.Run([]string{"ip-fetcher", "gcp"})
}

func TestGCPCmdEmptyPath(t *testing.T) {
	defer testCleanUp(os.Args)

	defer os.Unsetenv("TEST_EXIT")
	if os.Getenv("TEST_EXIT") == "1" {
		GCPCmdEmptyPath()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestGCPCmdEmptyPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestGCPCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.json"

	tDir := t.TempDir()

	_ = os.Setenv("IP_FETCHER_MOCK_GCP", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_GCP")

	app := getApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "gcp", "--path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "gcp", "--path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "cloud.json"))
}

func TestGCPCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	_ = os.Setenv("IP_FETCHER_MOCK_GCP", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_GCP")

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
	os.Args = []string{"ip-fetcher", "gcp", "--stdout"}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "35.219.128.0/24")
}

func TestGCPCmdStdOutAndFile(t *testing.T) {
	defer testCleanUp(os.Args)
	_ = os.Setenv("IP_FETCHER_MOCK_GCP", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_GCP")

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
	os.Args = []string{"ip-fetcher", "gcp", "--stdout", "--path", tDir}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "35.219.128.0/24")
	require.FileExists(t, filepath.Join(tDir, "cloud.json"))
}
