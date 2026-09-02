package main_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/agiledragon/gomonkey/v2"
	_ "github.com/agiledragon/gomonkey/v2/test/fake"
	mainpkg "github.com/jonhadfield/ip-fetcher/cmd/ip-fetcher"
	"github.com/stretchr/testify/require"
)

func GrafanaCmdNoStdOutNoPath() {
	app := mainpkg.GetApp()
	_ = app.Run([]string{"ip-fetcher", "grafana"})
}

func TestGrafanaCmdNoStdOutNoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	if os.Getenv("TEST_EXIT") == "1" {
		GrafanaCmdNoStdOutNoPath()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGrafanaCmdNoStdOutNoPath")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()

	e := &exec.ExitError{}
	if errors.As(err, &e) {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestGrafanaCmdSavetoPath(t *testing.T) {
	defer testCleanUp(os.Args)

	tDir := t.TempDir()
	t.Setenv("IP_FETCHER_MOCK_GRAFANA", "true")

	app := mainpkg.GetApp()

	os.Args = []string{"ip-fetcher", "grafana", "--Path", filepath.Join(tDir, "test.json")}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "test.json"))

	os.Args = []string{"ip-fetcher", "grafana", "--Path", tDir}
	require.NoError(t, app.Run(os.Args))
	require.FileExists(t, filepath.Join(tDir, "grafana.json"))
}

func TestGrafanaCmdStdOut(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_GRAFANA", "true")

	out := captureStdout(t, []string{"ip-fetcher", "grafana", "--stdout"})
	require.Contains(t, out, "40.176.0.202")
}

func TestGrafanaCmdLines(t *testing.T) {
	defer testCleanUp(os.Args)
	t.Setenv("IP_FETCHER_MOCK_GRAFANA", "true")

	tDir := t.TempDir()

	out := captureStdout(t, []string{"ip-fetcher", "grafana", "--stdout", "--lines", "--Path", tDir})
	require.Contains(t, out, "40.176.0.202")
	require.FileExists(t, filepath.Join(tDir, "grafana-prefixes.txt"))
}
