package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/cinsscore"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func cinsscoreCmd() *cli.Command {
	const (
		providerName  = "cinsscore"
		fileNameData  = "cinsscore.txt"
		fileNameLines = "cinsscore-prefixes.txt"
	)

	return &cli.Command{
		Name:         providerName,
		HelpName:     "- fetch CINS Army list addresses",
		Usage:        "CINS Army List (CI Army bad guys)",
		UsageText:    "ip-fetcher cinsscore {--stdout | --Path FILE} [--lines]",
		OnUsageError: onUsageError,
		Flags:        providerFlags(),
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			p := cinsscore.New()

			if isEnvEnabled("IP_FETCHER_MOCK_CINSSCORE") {
				defer gock.Off()

				u, _ := url.Parse(cinsscore.DownloadURL)
				gock.New(cinsscore.DownloadURL).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/cinsscore/testdata/ci-badguys.txt")
				gock.InterceptClient(p.Client.HTTPClient)
			}

			var data []byte
			if c.Bool(formatLines) {
				var doc cinsscore.Doc
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
