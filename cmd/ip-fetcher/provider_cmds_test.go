package main_test

import (
	"os"
	"path/filepath"
	"testing"

	mainpkg "github.com/jonhadfield/ip-fetcher/cmd/ip-fetcher"
	"github.com/stretchr/testify/require"
)

// these providers share one command shape, so a single table covers saving,
// stdout and --lines for each of them.
func TestProviderCmds(t *testing.T) {
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
		{"tenable", "IP_FETCHER_MOCK_TENABLE", "tenable.json", "tenable-prefixes.txt", "13.115.104.128"},
		{"detectify", "IP_FETCHER_MOCK_DETECTIFY", "detectify.txt", "detectify-prefixes.txt", "52.17.98.131"},
		{"tor", "IP_FETCHER_MOCK_TOR", "tor.txt", "tor-prefixes.txt", "185.220.101.34"},
		{"feodo", "IP_FETCHER_MOCK_FEODO", "feodo.txt", "feodo-prefixes.txt", "185.117.90.6"},
		{"okta", "IP_FETCHER_MOCK_OKTA", "okta.json", "okta-prefixes.txt", "35.247.69.17"},
		{"m365", "IP_FETCHER_MOCK_M365", "m365.json", "m365-prefixes.txt", "13.107.6.152"},
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
