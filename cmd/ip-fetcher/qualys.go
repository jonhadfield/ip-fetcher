package main

import (
	"fmt"
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/qualys"

	"github.com/urfave/cli/v2"
)

func qualysCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "qualys",
		helpName:  "Qualys AS27385 prefixes",
		usage:     "Qualys (AS27385 announced prefixes)",
		dataFile:  "qualys.json",
		linesFile: "qualys-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_QUALYS",
		mocks: []mockSource{
			{fmt.Sprintf(qualys.DownloadURL, qualys.ASNs[0]), "../../providers/qualys/testdata/prefixes.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := qualys.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
