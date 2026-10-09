package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/emergingthreats"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func emergingthreatsCmd() *cli.Command {
	const (
		providerName  = "emergingthreats"
		fileNameData  = "emergingthreats.txt"
		fileNameLines = "emergingthreats-prefixes.txt"
	)

	return &cli.Command{
		Name:         providerName,
		HelpName:     "- fetch Emerging Threats compromised addresses",
		Usage:        "Emerging Threats (compromised hosts)",
		UsageText:    "ip-fetcher emergingthreats {--stdout | --Path FILE} [--lines]",
		OnUsageError: onUsageError,
		Flags:        providerFlags(),
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			p := emergingthreats.New()

			if isEnvEnabled("IP_FETCHER_MOCK_EMERGINGTHREATS") {
				defer gock.Off()

				u, _ := url.Parse(emergingthreats.DownloadURL)
				gock.New(emergingthreats.DownloadURL).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/emergingthreats/testdata/compromised-ips.txt")
				gock.InterceptClient(p.Client.HTTPClient)
			}

			var data []byte
			if c.Bool(formatLines) {
				var doc emergingthreats.Doc
				if doc, err = p.Fetch(); err != nil {
					return err
				}

				if data, err = docToLinesWithCLI(c, doc); err != nil {
					return err
				}
			} else {
				data, _, _, err = p.FetchData()
				if err != nil {
					return err
				}
			}

			defaultName := fileNameData
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
