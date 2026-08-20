package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/akamai"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func akamaiCmd() *cli.Command {
	const (
		providerName = "akamai"
		fileName     = "prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Akamai prefixes",
		Usage:     "Akamai",
		UsageText: "ip-fetcher akamai {--stdout | --Path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)
			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagPath,
				Usage: usageWhereToSaveFile, Aliases: []string{"p"}, TakesFile: true,
			},
			&cli.BoolFlag{
				Name:  flagStdout,
				Usage: usageWriteToStdout, Aliases: []string{"s"},
			},
		},
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			a := akamai.New()

			if isEnvEnabled("IP_FETCHER_MOCK_AKAMAI") {
				defer gock.Off()
				urlBase := akamai.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/akamai/testdata/cidrs.zip")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			// Akamai publishes its CIDRs inside a zip, so the response has
			// to be parsed before it can be written as text.
			prefixes, err := a.Fetch()
			if err != nil {
				return err
			}

			data, err := docToLines(prefixes)
			if err != nil {
				return err
			}

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        providerName,
				DefaultFileName: fileName,
				Data:            data,
			})
		},
	}
}
