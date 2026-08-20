package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/spamhaus"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func spamhausCmd() *cli.Command {
	const (
		providerName  = "spamhaus"
		fileNameData  = "spamhaus.json"
		fileNameLines = "spamhaus-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Spamhaus DROP prefixes",
		Usage:     "Spamhaus DROP (Don't Route Or Peer)",
		UsageText: "ip-fetcher spamhaus {--stdout | --Path FILE} [--lines]",
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

			a := spamhaus.New()

			if isEnvEnabled("IP_FETCHER_MOCK_SPAMHAUS") {
				defer gock.Off()

				v4, _ := url.Parse(spamhaus.IPv4URL)
				gock.New(fmt.Sprintf("%s://%s", v4.Scheme, v4.Host)).
					Get(v4.Path).
					Reply(http.StatusOK).
					File("../../providers/spamhaus/testdata/drop_v4.json")

				v6, _ := url.Parse(spamhaus.IPv6URL)
				gock.New(fmt.Sprintf("%s://%s", v6.Scheme, v6.Host)).
					Get(v6.Path).
					Reply(http.StatusOK).
					File("../../providers/spamhaus/testdata/drop_v6.json")

				gock.InterceptClient(a.Client.HTTPClient)
			}

			var data []byte
			if c.Bool(formatLines) {
				var doc spamhaus.Doc
				if doc, err = a.Fetch(); err != nil {
					return err
				}

				if data, err = docToLines(doc); err != nil {
					return err
				}
			} else {
				data, _, _, err = a.FetchData()
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
