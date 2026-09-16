package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/intercom"

	"github.com/urfave/cli/v2"
)

func intercomCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "intercom",
		helpName:  "Intercom outbound addresses",
		usage:     "Intercom",
		dataFile:  "intercom.json",
		linesFile: "intercom-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_INTERCOM",
		mocks: []mockSource{
			{intercom.USURL, "../../providers/intercom/testdata/us.json"},
			{intercom.EUURL, "../../providers/intercom/testdata/eu.json"},
			{intercom.AUURL, "../../providers/intercom/testdata/au.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := intercom.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
