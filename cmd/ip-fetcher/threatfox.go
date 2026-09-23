package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/threatfox"

	"github.com/urfave/cli/v2"
)

func threatfoxCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "threatfox",
		helpName:  "abuse.ch ThreatFox recent C2 indicators",
		usage:     "abuse.ch ThreatFox (recent malware C2 ip:port indicators)",
		dataFile:  "threatfox.json",
		linesFile: "threatfox-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_THREATFOX",
		mocks: []mockSource{
			{threatfox.DownloadURL, "../../providers/threatfox/testdata/threatfox.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := threatfox.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
