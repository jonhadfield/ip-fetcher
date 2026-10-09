package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/digitalocean"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func digitaloceanCmd() *cli.Command {
	const (
		fileNameData  = "google.csv"
		fileNameLines = "digitalocean-prefixes.txt"
	)

	return &cli.Command{
		Name:         "digitalocean",
		HelpName:     "- fetch DigitalOcean prefixes",
		Usage:        "DigitalOcean",
		UsageText:    "ip-fetcher digitalocean {--stdout | --Path FILE} [--lines]",
		OnUsageError: onUsageError,
		Flags:        providerFlags(),
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			a := digitalocean.New()

			if isEnvEnabled("IP_FETCHER_MOCK_DIGITALOCEAN") {
				defer gock.Off()
				urlBase := digitalocean.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/digitalocean/testdata/google.csv")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			var data []byte
			if c.Bool(formatLines) {
				var doc digitalocean.Doc
				if doc, err = a.Fetch(); err != nil {
					return err
				}
				if data, err = docToLinesWithCLI(c, doc); err != nil {
					return err
				}
			} else {
				data, _, _, err = a.FetchData()
				if err != nil {
					return err
				}
			}

			defaultName := fileNameData
			if c.Bool(formatLines) {
				defaultName = fileNameLines
			}

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        "digitalocean",
				DefaultFileName: defaultName,
				Data:            data,
			})
		},
	}
}
