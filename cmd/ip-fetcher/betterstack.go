package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/betterstack"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func betterstackCmd() *cli.Command {
	const (
		providerName  = "betterstack"
		fileNameData  = "betterstack.txt"
		fileNameLines = "betterstack-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Better Stack probe addresses",
		Usage:     "Better Stack (uptime check probes)",
		UsageText: "ip-fetcher betterstack {--stdout | --Path FILE} [--lines]",
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

			p := betterstack.New()

			if isEnvEnabled("IP_FETCHER_MOCK_BETTERSTACK") {
				defer gock.Off()

				u, _ := url.Parse(betterstack.DownloadURL)
				gock.New(betterstack.DownloadURL).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/betterstack/testdata/ips.txt")
				gock.InterceptClient(p.Client.HTTPClient)
			}

			data, err := betterstackData(c, &p)
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

// betterstackData returns newline separated prefixes when --lines is set, and the
// upstream document otherwise.
func betterstackData(c *cli.Context, p *betterstack.BetterStack) ([]byte, error) {
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
