package main_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/agiledragon/gomonkey/v2"
	_ "github.com/agiledragon/gomonkey/v2/test/fake"
	mainpkg "github.com/jonhadfield/ip-fetcher/cmd/ip-fetcher"
	"github.com/stretchr/testify/require"
)

func ICloudPRCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", mainpkg.SICloudPR})
}

func TestICloudPRCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		ICloudPRCmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestICloudPRCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	var e *exec.ExitError
	if errors.As(err, &e) && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func ICloudPRCmdEmptyPath() {
	defer testCleanUp(os.Args)

	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", mainpkg.SICloudPR})
}

func TestICloudPRCmdEmptyPath(t *testing.T) {
	defer testCleanUp(os.Args)

	defer os.Unsetenv("TEST_EXIT")
	if os.Getenv("TEST_EXIT") == "1" {
		ICloudPRCmdEmptyPath()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestICloudPRCmdEmptyPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestICloudPRCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "prefixes.csv"

	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_ICLOUDPR", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_ICLOUDPR")

	app := mainpkg.GetApp()

	// with filename only
	os.Args = []string{"ip-fetcher", mainpkg.SICloudPR, "--Path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", mainpkg.SICloudPR, "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "prefixes.csv"))
}

func TestICloudPRCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_ICLOUDPR", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_ICLOUDPR")

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
	os.Args = []string{"ip-fetcher", mainpkg.SICloudPR, "--stdout"}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC

	require.Contains(t, out, "172.224.224.40/31")
}

func TestICloudPRCmdStdOutAndFile(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_ICLOUDPR", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_ICLOUDPR")

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

	app := mainpkg.GetApp()
	os.Args = []string{"ip-fetcher", mainpkg.SICloudPR, "--stdout", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "172.224.224.32/31")
	require.FileExists(t, filepath.Join(tDir, "prefixes.csv"))
}
