package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/x4bnet"

	"github.com/urfave/cli/v2"
)

func x4bnetCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "x4bnet",
		helpName:  "X4BNet commercial VPN prefixes",
		usage:     "X4BNet VPN",
		dataFile:  "x4bnet.json",
		linesFile: "x4bnet-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_X4BNET",
		mocks: []mockSource{
			{x4bnet.IPv4URL, "../../providers/x4bnet/testdata/ipv4.txt"},
			{x4bnet.IPv6URL, "../../providers/x4bnet/testdata/ipv6.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := x4bnet.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
