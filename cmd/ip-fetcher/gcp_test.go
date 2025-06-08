package main_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	_ "github.com/agiledragon/gomonkey/v2"
	_ "github.com/agiledragon/gomonkey/v2/test/fake"
	mainpkg "github.com/jonhadfield/ip-fetcher/cmd/ip-fetcher"
	"github.com/stretchr/testify/require"
)

func GCPCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
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
	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func GCPCmdEmptyPath() {
	defer testCleanUp(os.Args)

	app := mainpkg.GetApp()
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
	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestGCPCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.json"

	tDir := t.TempDir()

	t.Setenv("IP_FETCHER_MOCK_GCP", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_GCP")

	app := mainpkg.GetApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "gcp", "--Path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "gcp", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "cloud.json"))
}

func TestGCPCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_GCP", "true")
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

	app := mainpkg.GetApp()
	os.Args = []string{"ip-fetcher", "gcp", "--stdout"}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "35.219.128.0/24")
}

func TestGCPCmdStdOutAndFile(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_GCP", "true")
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

	app := mainpkg.GetApp()
	os.Args = []string{"ip-fetcher", "gcp", "--stdout", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "35.219.128.0/24")
	require.FileExists(t, filepath.Join(tDir, "cloud.json"))
}

func TestGCPCmdStdOutLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_GCP", "true")
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

	app := mainpkg.GetApp()
	os.Args = []string{"ip-fetcher", "gcp", "--stdout", "--format", "lines", tDir}
	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC
	require.Len(t, regexp.MustCompile(`35\.219\.128\.0/24(\r\n|\r|\n)`).FindAllString(out, -1), 1)
	require.NoFileExists(t, filepath.Join(tDir, "cloud.json"))
}
