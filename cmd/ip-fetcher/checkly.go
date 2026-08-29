package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/checkly"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func checklyCmd() *cli.Command {
	const (
		providerName  = "checkly"
		fileNameData  = "checkly.json"
		fileNameLines = "checkly-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Checkly probe addresses",
		Usage:     "Checkly (monitoring probe addresses)",
		UsageText: "ip-fetcher checkly {--stdout | --Path FILE} [--lines]",
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

			p := checkly.New()

			if isEnvEnabled("IP_FETCHER_MOCK_CHECKLY") {
				defer gock.Off()

				for _, l := range []struct{ raw, file string }{
					{checkly.IPv4URL, "../../providers/checkly/testdata/static-ips.json"},
					{checkly.IPv6URL, "../../providers/checkly/testdata/static-ipv6s.json"},
				} {
					u, _ := url.Parse(l.raw)
					gock.New(l.raw).
						Get(u.Path).
						Reply(http.StatusOK).
						File(l.file)
				}

				gock.InterceptClient(p.Client.HTTPClient)
			}

			data, err := checklyData(c, &p)
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

// checklyData returns newline separated prefixes when --lines is set, and the
// upstream document otherwise.
func checklyData(c *cli.Context, p *checkly.Checkly) ([]byte, error) {
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
