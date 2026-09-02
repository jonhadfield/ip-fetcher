package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/site24x7"

	"github.com/urfave/cli/v2"
)

func site24x7Cmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "site24x7",
		helpName:  "Site24x7 monitoring location addresses",
		usage:     "Site24x7 (monitoring locations)",
		dataFile:  "site24x7.json",
		linesFile: "site24x7-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_SITE24X7",
		mocks: []mockSource{
			{site24x7.LocationsURL, "../../providers/site24x7/testdata/locations.html"},
			// the export link the fixture page carries.
			{
				"https://creatorapp.zohopublic.in/mesite24x7/location-manager/json/IP_Address_View/TESTTOKEN",
				"../../providers/site24x7/testdata/locations.json",
			},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := site24x7.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
