package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/cymru"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func cymruCmd() *cli.Command {
	const (
		providerName  = "cymru"
		fileNameData  = "cymru.json"
		fileNameLines = "cymru-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Team Cymru bogon prefixes",
		Usage:     "Team Cymru Bogons (unallocated and reserved prefixes)",
		UsageText: "ip-fetcher cymru {--stdout | --Path FILE} [--lines]",
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

			p := cymru.New()

			if isEnvEnabled("IP_FETCHER_MOCK_CYMRU") {
				defer gock.Off()

				for _, l := range []struct{ raw, file string }{
					{cymru.IPv4URL, "../../providers/cymru/testdata/fullbogons-ipv4.txt"},
					{cymru.IPv6URL, "../../providers/cymru/testdata/fullbogons-ipv6.txt"},
				} {
					u, _ := url.Parse(l.raw)
					gock.New(l.raw).
						Get(u.Path).
						Reply(http.StatusOK).
						File(l.file)
				}

				gock.InterceptClient(p.Client.HTTPClient)
			}

			var data []byte
			if c.Bool(formatLines) {
				var doc cymru.Doc
				if doc, err = p.Fetch(); err != nil {
					return err
				}

				if data, err = docToLines(doc); err != nil {
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
