package main

import (
	"fmt"
	"net/http"
	"net/url"

	output2 "github.com/jonhadfield/ip-fetcher/internal/output"

	"github.com/jonhadfield/ip-fetcher/providers/zscaler"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func zscalerCmd() *cli.Command {
	const (
		providerName = "zscaler"
		fileName     = "prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Zscaler prefixes",
		Usage:     "Zscaler",
		UsageText: "ip-fetcher zscaler {--stdout | --Path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)
			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "Path",
				Usage: usageWhereToSaveFile, Aliases: []string{"p"}, TakesFile: true,
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: usageWriteToStdout, Aliases: []string{"s"},
			},
		},
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			z := zscaler.New()

			if isEnvEnabled("IP_FETCHER_MOCK_ZSCALER") {
				defer gock.Off()
				urlBase := zscaler.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/zscaler/testdata/doc.json")
				gock.InterceptClient(z.Client.HTTPClient)
			}

			data, _, _, err := z.FetchData()
			if err != nil {
				return err
			}

			if stdout {
				if err = output2.PrettyPrintJSON(data); err != nil {
					return fmt.Errorf("error printing data to stdout: %w", err)
				}
			}

			return writeOutputs(path, false, SaveFileInput{
				Provider:        providerName,
				DefaultFileName: fileName,
				Data:            data,
			})
		},
	}
}
