package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/telegram"

	"github.com/urfave/cli/v2"
)

func telegramCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "telegram",
		helpName:  "Telegram DC and Bot API prefixes",
		usage:     "Telegram",
		dataFile:  "telegram.txt",
		linesFile: "telegram-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_TELEGRAM",
		mocks: []mockSource{
			{telegram.DownloadURL, "../../providers/telegram/testdata/telegram.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := telegram.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
