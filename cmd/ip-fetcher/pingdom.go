package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/pingdom"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func pingdomCmd() *cli.Command {
	const (
		providerName  = "pingdom"
		fileNameData  = "pingdom.json"
		fileNameLines = "pingdom-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Pingdom probe addresses",
		Usage:     "Pingdom (monitoring probe servers)",
		UsageText: "ip-fetcher pingdom {--stdout | --Path FILE} [--lines]",
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

			p := pingdom.New()

			if isEnvEnabled("IP_FETCHER_MOCK_PINGDOM") {
				defer gock.Off()

				for _, l := range []struct{ raw, file string }{
					{pingdom.IPv4URL, "../../providers/pingdom/testdata/probes-ipv4.txt"},
					{pingdom.IPv6URL, "../../providers/pingdom/testdata/probes-ipv6.txt"},
				} {
					u, _ := url.Parse(l.raw)
					gock.New(l.raw).
						Get(u.Path).
						Reply(http.StatusOK).
						File(l.file)
				}

				gock.InterceptClient(p.Client.HTTPClient)
			}

			data, err := pingdomData(c, &p)
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

// pingdomData returns newline separated prefixes when --lines is set, and the
// upstream document otherwise.
func pingdomData(c *cli.Context, p *pingdom.Pingdom) ([]byte, error) {
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
