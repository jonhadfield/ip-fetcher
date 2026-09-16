package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/amazonbot"

	"github.com/urfave/cli/v2"
)

func amazonbotCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "amazonbot",
		helpName:  "Amazonbot, Amzn-SearchBot and Amzn-User prefixes",
		usage:     "Amazonbot",
		dataFile:  "amazonbot.json",
		linesFile: "amazonbot-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_AMAZONBOT",
		mocks: []mockSource{
			{amazonbot.AmazonbotURL, "../../providers/amazonbot/testdata/amazonbot.html"},
			{amazonbot.SearchBotURL, "../../providers/amazonbot/testdata/searchbot.html"},
			{amazonbot.UserURL, "../../providers/amazonbot/testdata/user.html"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := amazonbot.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
