package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/sentry"

	"github.com/urfave/cli/v2"
)

func sentryCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "sentry",
		helpName:  "Sentry uptime check addresses",
		usage:     "Sentry Uptime",
		dataFile:  "sentry.txt",
		linesFile: "sentry-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_SENTRY",
		mocks: []mockSource{
			{sentry.DownloadURL, "../../providers/sentry/testdata/sentry.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := sentry.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
