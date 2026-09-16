package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/salesforce"

	"github.com/urfave/cli/v2"
)

func salesforceCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "salesforce",
		helpName:  "Salesforce Hyperforce prefixes",
		usage:     "Salesforce",
		dataFile:  "salesforce.json",
		linesFile: "salesforce-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_SALESFORCE",
		mocks: []mockSource{
			{salesforce.DownloadURL, "../../providers/salesforce/testdata/salesforce.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := salesforce.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
