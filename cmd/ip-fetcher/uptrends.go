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
		Name:      providerName,
		HelpName:  "- fetch Uptrends checkpoint addresses",
		Usage:     "Uptrends (monitoring checkpoints)",
		UsageText: "ip-fetcher uptrends {--stdout | --Path FILE} [--lines]",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagPath,
				Usage: usageWhereToSaveFile, Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  flagStdout,
				Usage: usageWriteToStdout, Aliases: []string{"s"},
			},
			&cli.BoolFlag{
				Name:  formatLines,
				Usage: usageLinesOutput,
			},
		},
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

			data, err := uptrendsData(c, &p)
			if err != nil {
				return err
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

// uptrendsData returns newline separated prefixes when --lines is set, and the
// upstream document otherwise.
func uptrendsData(c *cli.Context, p *uptrends.Uptrends) ([]byte, error) {
	if c.Bool(formatLines) {
		doc, err := p.Fetch()
		if err != nil {
			return nil, err
		}

		return docToLines(doc)
	}

	data, _, _, err := p.FetchData()

	return data, err
}
