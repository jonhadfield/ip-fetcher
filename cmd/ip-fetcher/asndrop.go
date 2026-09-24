package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/asndrop"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func asndropCmd() *cli.Command {
	const (
		providerName  = "asndrop"
		fileName      = "asndrop.json"
		fileNameLines = "asndrop-asns.txt"
	)

	return &cli.Command{
		Name:         providerName,
		HelpName:     "- fetch Spamhaus ASN-DROP",
		Usage:        "Spamhaus ASN-DROP (criminal and hijacked ASNs)",
		UsageText:    "ip-fetcher asndrop {--stdout | --Path FILE} [--lines]",
		OnUsageError: onUsageError,
		Flags:        providerFlags(),
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			a := asndrop.New()

			if isEnvEnabled("IP_FETCHER_MOCK_ASNDROP") {
				defer gock.Off()

				u, _ := url.Parse(asndrop.DownloadURL)
				gock.New(asndrop.DownloadURL).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/asndrop/testdata/asndrop.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			var data []byte
			if c.Bool(formatLines) {
				doc, fetchErr := a.Fetch()
				if fetchErr != nil {
					return fetchErr
				}

				data = doc.Lines()
			} else {
				data, _, _, err = a.FetchData()
				if err != nil {
					return err
				}
			}

			defaultName := fileName
			if c.Bool(formatLines) {
				defaultName = fileNameLines
			}

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        providerName,
				DefaultFileName: defaultName,
				Data:            data,
			})
		},
	}
}
