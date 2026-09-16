package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/ccbot"

	"github.com/urfave/cli/v2"
)

func ccbotCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "ccbot",
		helpName:  "Common Crawl CCBot prefixes",
		usage:     "Common Crawl CCBot",
		dataFile:  "ccbot.json",
		linesFile: "ccbot-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_CCBOT",
		mocks: []mockSource{
			{ccbot.DownloadURL, "../../providers/ccbot/testdata/ccbot.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := ccbot.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
