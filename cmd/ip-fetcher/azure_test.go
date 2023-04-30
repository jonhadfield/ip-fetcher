package main

import (
	"bytes"
	"github.com/jonhadfield/ip-fetcher/providers/azure"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const testAzureDownloadUrl = "https://download.microsoft.com/download/7/1/D/71D86715-5596-4529-9B13-DA13A5DE5B63/ServiceTags_Public_20221212.json"

func AzureCmdNoStdOutNoPath() {
	app := getApp()
	_ = app.Run([]string{"ip-fetcher", "azure"})
}

func TestAzureCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	defer os.Unsetenv("TEST_EXIT")
	if os.Getenv("TEST_EXIT") == "1" {
		AzureCmdNoStdOutNoPath()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestAzureCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestAzureCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	testFile := "test.json"

	tDir := t.TempDir()

	_ = os.Setenv("IP_FETCHER_MOCK_AZURE", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_AZURE")

	ac := azure.New()
	ac.DownloadURL = testAzureDownloadUrl

	app := getApp()

	// with filename only
	os.Args = []string{"ip-fetcher", "azure", "--path", filepath.Join(tDir, testFile)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, testFile))

	// with directory only
	os.Args = []string{"ip-fetcher", "azure", "--path", filepath.Join(tDir)}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "ServiceTags_Public.json"))
}

func TestAzureCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)

	_ = os.Setenv("IP_FETCHER_MOCK_AZURE", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_AZURE")

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
	os.Args = []string{"ip-fetcher", "azure", "--stdout"}

	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC

	require.Contains(t, out, "\"20.38.149.132/30\",")
}

func TestAzureCmdStdOutAndFile(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()

	_ = os.Setenv("IP_FETCHER_MOCK_AZURE", "true")
	defer os.Unsetenv("IP_FETCHER_MOCK_AZURE")

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

	app := getApp()
	os.Args = []string{"ip-fetcher", "azure", "--stdout", "--path", tDir}

	require.NoError(t, app.Run(os.Args))

	_ = w.Close()
	os.Stdout = old
	out := <-outC

	require.Contains(t, out, "\"20.38.149.132/30\",")
	require.FileExists(t, filepath.Join(tDir, "ServiceTags_Public.json"))
}
