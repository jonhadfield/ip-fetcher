package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/stopforumspam"

	"github.com/urfave/cli/v2"
)

func stopforumspamCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "stopforumspam",
		helpName:  "StopForumSpam toxic network prefixes",
		usage:     "StopForumSpam",
		dataFile:  "stopforumspam.txt",
		linesFile: "stopforumspam-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_STOPFORUMSPAM",
		mocks: []mockSource{
			{stopforumspam.DownloadURL, "../../providers/stopforumspam/testdata/toxic.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := stopforumspam.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
