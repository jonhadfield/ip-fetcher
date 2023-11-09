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

func GooglebotCmdNoStdOutNoPath() {
	app := getApp()
	_ = app.Run([]string{"ip-fetcher", "googlebot"})
}

func TestGooglebotCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		GooglebotCmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGooglebotCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func GooglebotCmdEmptyPath() {
	defer testCleanUp(os.Args)

	app := getApp()
	_ = app.Run([]string{"ip-fetcher", "googlebot"})
}

func TestGooglebotCmdEmptyPath(t *testing.T) {
	defer testCleanUp(os.Args)

	defer os.Unsetenv("TEST_EXIT")
	if os.Getenv("TEST_EXIT") == "1" {
		GooglebotCmdEmptyPath()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestGooglebotCmdEmptyPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestGooglebotCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.json"

	tDir := t.TempDir()

	_ = os.Setenv("IP_FETCHER_MOCK_GOOGLEBOT", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_GOOGLEBOT")

	app := getApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "googlebot", "--path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "googlebot", "--path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "googlebot.json"))
}

func TestGooglebotCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	_ = os.Setenv("IP_FETCHER_MOCK_GOOGLEBOT", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_GOOGLEBOT")

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
	os.Args = []string{"ip-fetcher", "googlebot", "--stdout"}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "2001:4860:4801:12::/64")
}

func TestGooglebotCmdStdOutAndFile(t *testing.T) {
	defer testCleanUp(os.Args)
	_ = os.Setenv("IP_FETCHER_MOCK_GOOGLEBOT", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_GOOGLEBOT")

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
	os.Args = []string{"ip-fetcher", "googlebot", "--stdout", "--path", tDir}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "66.249.79.32/27")
	require.Contains(t, out, "2001:4860:4801:12::/64")
	require.FileExists(t, filepath.Join(tDir, "googlebot.json"))
}
