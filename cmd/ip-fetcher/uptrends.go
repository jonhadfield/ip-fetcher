package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/uptrends"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func uptrendsCmd() *cli.Command {
	const (
		providerName  = "uptrends"
		fileNameData  = "uptrends.json"
		fileNameLines = "uptrends-prefixes.txt"
	)

	return &cli.Command{
		Name:         providerName,
		HelpName:     "- fetch Uptrends checkpoint addresses",
		Usage:        "Uptrends (monitoring checkpoints)",
		UsageText:    "ip-fetcher uptrends {--stdout | --Path FILE} [--lines]",
		OnUsageError: onUsageError,
		Flags:        providerFlags(),
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			p := uptrends.New()

			if isEnvEnabled("IP_FETCHER_MOCK_UPTRENDS") {
				defer gock.Off()

				for _, l := range []struct{ raw, file string }{
					{uptrends.IPv4URL, "../../providers/uptrends/testdata/ipv4.json"},
					{uptrends.IPv6URL, "../../providers/uptrends/testdata/ipv6.json"},
				} {
					u, _ := url.Parse(l.raw)
					gock.New(l.raw).
						Get(u.Path).
						Reply(http.StatusOK).
						File(l.file)
				}

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
