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
		{"amazonbot", "IP_FETCHER_MOCK_AMAZONBOT", "amazonbot.json", "amazonbot-prefixes.txt", "3.81.245.78"},
		{"ccbot", "IP_FETCHER_MOCK_CCBOT", "ccbot.json", "ccbot-prefixes.txt", "3.41.188.32"},
		{"cachefly", "IP_FETCHER_MOCK_CACHEFLY", "cachefly.txt", "cachefly-prefixes.txt", "205.234.175.0"},
		{"salesforce", "IP_FETCHER_MOCK_SALESFORCE", "salesforce.json", "salesforce-prefixes.txt", "145.224.193.0"},
		{"huawei", "IP_FETCHER_MOCK_HUAWEI", "huawei.json", "huawei-prefixes.txt", "192.0.2.0"},
		{"mullvad", "IP_FETCHER_MOCK_MULLVAD", "mullvad.json", "mullvad-prefixes.txt", "103.124.165.2"},
		{"gitlab", "IP_FETCHER_MOCK_GITLAB", "gitlab.txt", "gitlab-prefixes.txt", "34.74.90.64"},
		{"intercom", "IP_FETCHER_MOCK_INTERCOM", "intercom.json", "intercom-prefixes.txt", "34.197.76.213"},
		{"quiccloud", "IP_FETCHER_MOCK_QUICCLOUD", "quiccloud.txt", "quiccloud-prefixes.txt", "102.221.36.98"},
		{"telegram", "IP_FETCHER_MOCK_TELEGRAM", "telegram.txt", "telegram-prefixes.txt", "91.108.56.0"},
		{"circleci", "IP_FETCHER_MOCK_CIRCLECI", "circleci.json", "circleci-prefixes.txt", "100.27.248.128"},
		{"threatfox", "IP_FETCHER_MOCK_THREATFOX", "threatfox.json", "threatfox-prefixes.txt", "155.103.69.239"},
		{"binarydefense", "IP_FETCHER_MOCK_BINARYDEFENSE", "binarydefense.txt", "binarydefense-prefixes.txt", "1.20.168.127"},
		{"ipsum", "IP_FETCHER_MOCK_IPSUM", "ipsum.txt", "ipsum-prefixes.txt", "94.154.43.254"},
		{"hetrixtools", "IP_FETCHER_MOCK_HETRIXTOOLS", "hetrixtools.txt", "hetrixtools-prefixes.txt", "52.207.41.187"},
		{"nodeping", "IP_FETCHER_MOCK_NODEPING", "nodeping.txt", "nodeping-prefixes.txt", "104.247.192.170"},
		{"qualys", "IP_FETCHER_MOCK_QUALYS", "qualys.json", "qualys-prefixes.txt", "64.39.96.0"},
		{"airvpn", "IP_FETCHER_MOCK_AIRVPN", "airvpn.json", "airvpn-prefixes.txt", "185.156.175.170"},
		{"ivpn", "IP_FETCHER_MOCK_IVPN", "ivpn.json", "ivpn-prefixes.txt", "149.22.83.100"},
		{"surfshark", "IP_FETCHER_MOCK_SURFSHARK", "surfshark.json", "surfshark-prefixes.txt", "172.216.15.93"},
		{"x4bnet", "IP_FETCHER_MOCK_X4BNET", "x4bnet.json", "x4bnet-prefixes.txt", "2.26.157.0"},
		{"stopforumspam", "IP_FETCHER_MOCK_STOPFORUMSPAM", "stopforumspam.txt", "stopforumspam-prefixes.txt", "103.81.182.0"},
		{"asndrop", "IP_FETCHER_MOCK_ASNDROP", "asndrop.json", "asndrop-asns.txt", "245"},
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
