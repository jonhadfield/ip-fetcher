package main

import (
	"context"
	"net/http"
	"net/netip"

	"github.com/jonhadfield/ip-fetcher/providers/surfshark"

	"github.com/urfave/cli/v2"
)

func surfsharkCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "surfshark",
		helpName:  "Surfshark VPN egress addresses",
		usage:     "Surfshark",
		dataFile:  "surfshark.json",
		linesFile: "surfshark-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_SURFSHARK",
		mocks: []mockSource{
			{surfshark.DownloadURL, "../../providers/surfshark/testdata/clusters.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := surfshark.New()
			if isEnvEnabled("IP_FETCHER_MOCK_SURFSHARK") {
				p.LookupIP = func(_ context.Context, host string) ([]netip.Addr, error) {
					switch host {
					case "al-tia.prod.surfshark.com":
						return []netip.Addr{netip.MustParseAddr("172.216.15.93")}, nil
					case "uk-lon.prod.surfshark.com":
						return []netip.Addr{
							netip.MustParseAddr("138.199.29.230"),
							netip.MustParseAddr("2a01:db8::1"),
						}, nil
					default:
						return nil, nil
					}
				}
			}

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
