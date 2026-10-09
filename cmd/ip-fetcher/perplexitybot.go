package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/perplexitybot"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func perplexitybotCmd() *cli.Command {
	const (
		providerName  = "perplexitybot"
		fileName      = "perplexitybot.json"
		fileNameLines = "perplexitybot-prefixes.txt"
	)

	return &cli.Command{
		Name:         providerName,
		HelpName:     "- fetch PerplexityBot prefixes",
		Usage:        "Perplexity's web crawler",
		UsageText:    "ip-fetcher perplexitybot {--stdout | --Path FILE}",
		OnUsageError: onUsageError,
		Flags:        providerFlags(),
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			a := perplexitybot.New()

			if isEnvEnabled("IP_FETCHER_MOCK_PERPLEXITYBOT") {
				defer gock.Off()
				urlBase := perplexitybot.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/perplexitybot/testdata/perplexitybot.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			var data []byte
			if c.Bool(formatLines) {
				var doc perplexitybot.Doc
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
