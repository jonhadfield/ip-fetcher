package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/site24x7"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func site24x7Cmd() *cli.Command {
	const (
		providerName  = "site24x7"
		fileNameData  = "site24x7.json"
		fileNameLines = "site24x7-prefixes.txt"
	)

	return &cli.Command{
		Name:         providerName,
		HelpName:     "- fetch Site24x7 monitoring location addresses",
		Usage:        "Site24x7 (monitoring locations)",
		UsageText:    "ip-fetcher site24x7 {--stdout | --Path FILE} [--lines]",
		OnUsageError: onUsageError,
		Flags:        providerFlags(),
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			p := site24x7.New()

			if isEnvEnabled("IP_FETCHER_MOCK_SITE24X7") {
				defer gock.Off()

				u, _ := url.Parse(site24x7.LocationsURL)
				gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/site24x7/testdata/locations.html")

				// the export link the fixture page carries.
				gock.New("https://creatorapp.zohopublic.in").
					Get("/mesite24x7/location-manager/json/IP_Address_View/TESTTOKEN").
					Reply(http.StatusOK).
					File("../../providers/site24x7/testdata/locations.json")

				gock.InterceptClient(p.Client.HTTPClient)
			}

			data, err := providerData(c, p.FetchData, func() (any, error) { return p.Fetch() })
			if err != nil {
				return err
			}

			return writeProviderOutputs(c, path, stdout, providerName, fileNameData, fileNameLines, data)
		},
	}
}
