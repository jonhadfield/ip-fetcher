package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/updown"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func updownCmd() *cli.Command {
	const (
		providerName  = "updown"
		fileNameData  = "updown.json"
		fileNameLines = "updown-prefixes.txt"
	)

	return &cli.Command{
		Name:         providerName,
		HelpName:     "- fetch updown.io monitoring node addresses",
		Usage:        "updown.io (monitoring nodes)",
		UsageText:    "ip-fetcher updown {--stdout | --Path FILE} [--lines]",
		OnUsageError: onUsageError,
		Flags:        providerFlags(),
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			p := updown.New()

			if isEnvEnabled("IP_FETCHER_MOCK_UPDOWN") {
				defer gock.Off()

				u, _ := url.Parse(updown.DownloadURL)
				gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/updown/testdata/nodes.json")

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
