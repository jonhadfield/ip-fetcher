package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/sentry"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func sentryCmd() *cli.Command {
	const (
		providerName  = "sentry"
		fileNameData  = "sentry.txt"
		fileNameLines = "sentry-prefixes.txt"
	)

	return &cli.Command{
		Name:         providerName,
		HelpName:     "- fetch Sentry uptime check addresses",
		Usage:        "Sentry Uptime",
		UsageText:    "ip-fetcher sentry {--stdout | --Path FILE} [--lines]",
		OnUsageError: onUsageError,
		Flags:        providerFlags(),
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			p := sentry.New()

			if isEnvEnabled("IP_FETCHER_MOCK_SENTRY") {
				defer gock.Off()

				u, _ := url.Parse(sentry.DownloadURL)
				gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/sentry/testdata/sentry.txt")

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
