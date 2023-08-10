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

func OCICmdNoStdOutNoPath() {
	app := getApp()
	_ = app.Run([]string{"ip-fetcher", "oci"})
}

func TestOCICmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		OCICmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestOCICmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func OCICmdEmptyPath() {
	defer testCleanUp(os.Args)

	app := getApp()
	_ = app.Run([]string{"ip-fetcher", "oci"})
}

func TestOCICmdEmptyPath(t *testing.T) {
	defer testCleanUp(os.Args)

	defer os.Unsetenv("TEST_EXIT")
	if os.Getenv("TEST_EXIT") == "1" {
		OCICmdEmptyPath()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestOCICmdEmptyPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestOCICmdSavetoPathFileNameOnly(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.json"

	tDir := t.TempDir()

	_ = os.Setenv("IP_FETCHER_MOCK_OCI", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_OCI")

	app := getApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "oci", "--path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))
}

func TestOCICmdSavetoPathDirectoryOnly(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()

	_ = os.Setenv("IP_FETCHER_MOCK_OCI", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_OCI")

	app := getApp()

	// with directory only
	os.Args = []string{"ip-fetcher", "oci", "--path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "public_ip_ranges.json"))
}

func TestOCICmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	_ = os.Setenv("IP_FETCHER_MOCK_OCI", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_OCI")

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
	os.Args = []string{"ip-fetcher", "oci", "--stdout"}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "192.29.172.1/25")
}

func TestOCICmdStdOutAndFile(t *testing.T) {
	defer testCleanUp(os.Args)
	_ = os.Setenv("IP_FETCHER_MOCK_OCI", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_OCI")

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
	os.Args = []string{"ip-fetcher", "oci", "--stdout", "--path", tDir}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "192.29.160.1/21")
	require.FileExists(t, filepath.Join(tDir, "public_ip_ranges.json"))
}
