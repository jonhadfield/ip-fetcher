package main

import (
	"fmt"
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/huawei"

	"github.com/urfave/cli/v2"
)

func huaweiCmd() *cli.Command {
	mocks := make([]mockSource, 0, len(huawei.ASNs))
	for _, asn := range huawei.ASNs {
		mocks = append(mocks, mockSource{
			fmt.Sprintf(huawei.DownloadURL, asn),
			"../../providers/huawei/testdata/prefixes.json",
		})
	}

	return providerCommand(providerSpec{
		name:      "huawei",
		helpName:  "Huawei Cloud prefixes",
		usage:     "Huawei Cloud",
		dataFile:  "huawei.json",
		linesFile: "huawei-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_HUAWEI",
		mocks:     mocks,
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := huawei.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
