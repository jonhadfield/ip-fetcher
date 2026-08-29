package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/statuscake"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func statuscakeCmd() *cli.Command {
	const (
		providerName  = "statuscake"
		fileNameData  = "statuscake.json"
		fileNameLines = "statuscake-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch StatusCake probe addresses",
		Usage:     "StatusCake (monitoring probe locations)",
		UsageText: "ip-fetcher statuscake {--stdout | --Path FILE} [--lines]",
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

			p := statuscake.New()

			if isEnvEnabled("IP_FETCHER_MOCK_STATUSCAKE") {
				defer gock.Off()

				u, _ := url.Parse(statuscake.DownloadURL)
				gock.New(statuscake.DownloadURL).
					Get(u.Path).
					MatchParam("format", "json").
					Reply(http.StatusOK).
					File("../../providers/statuscake/testdata/locations.json")
				gock.InterceptClient(p.Client.HTTPClient)
			}

			data, err := statuscakeData(c, &p)
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

// statuscakeData returns newline separated prefixes when --lines is set, and the
// upstream document otherwise.
func statuscakeData(c *cli.Context, p *statuscake.StatusCake) ([]byte, error) {
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
