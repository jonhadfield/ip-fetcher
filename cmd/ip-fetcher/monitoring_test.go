package main_test

import (
	"os"
	"path/filepath"
	"testing"

	mainpkg "github.com/jonhadfield/ip-fetcher/cmd/ip-fetcher"
	"github.com/stretchr/testify/require"
)

// the five monitoring providers added together share one command shape, so a
// single table covers saving, stdout and --lines for each of them.
func TestMonitoringProviderCmds(t *testing.T) {
	cases := []struct {
		provider  string
		mockEnv   string
		dataFile  string
		linesFile string
		address   string
	}{
		{"grafana", "IP_FETCHER_MOCK_GRAFANA", "grafana.json", "grafana-prefixes.txt", "40.176.0.202"},
		{"sentry", "IP_FETCHER_MOCK_SENTRY", "sentry.txt", "sentry-prefixes.txt", "34.123.33.225"},
		{"site24x7", "IP_FETCHER_MOCK_SITE24X7", "site24x7.json", "site24x7-prefixes.txt", "37.221.111.107"},
		{"updown", "IP_FETCHER_MOCK_UPDOWN", "updown.json", "updown-prefixes.txt", "45.32.74.41"},
		{"uptrends", "IP_FETCHER_MOCK_UPTRENDS", "uptrends.json", "uptrends-prefixes.txt", "101.201.208.194"},
	}

	for _, tc := range cases {
		t.Run(tc.provider, func(t *testing.T) {
			defer testCleanUp(os.Args)

			t.Setenv(tc.mockEnv, "true")

			tDir := t.TempDir()
			app := mainpkg.GetApp()

			// a named file is used as given.
			os.Args = []string{"ip-fetcher", tc.provider, "--Path", filepath.Join(tDir, "named.out")}
			require.NoError(t, app.Run(os.Args))
			require.FileExists(t, filepath.Join(tDir, "named.out"))

			// a directory takes the provider's own file name.
			os.Args = []string{"ip-fetcher", tc.provider, "--Path", tDir}
			require.NoError(t, app.Run(os.Args))
			require.FileExists(t, filepath.Join(tDir, tc.dataFile))

			out := captureStdout(t, []string{"ip-fetcher", tc.provider, "--stdout"})
			require.Contains(t, out, tc.address)

			// --lines writes prefixes, under the prefixes file name.
			out = captureStdout(t, []string{"ip-fetcher", tc.provider, "--stdout", "--lines", "--Path", tDir})
			require.Contains(t, out, tc.address)
			require.FileExists(t, filepath.Join(tDir, tc.linesFile))
		})
	}
}
