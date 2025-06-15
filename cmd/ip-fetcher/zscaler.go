package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

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
			path := strings.TrimSpace(c.String("Path"))
			if path == "" && !c.Bool("stdout") {
				_ = cli.ShowSubcommandHelp(c)
				fmt.Println("\n" + errStdoutOrPathRequired)
				os.Exit(1)
			}

			z := zscaler.New()

			if os.Getenv("IP_FETCHER_MOCK_ZSCALER") == "true" {
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

			if path != "" {
				var out string
				out, err = SaveFile(SaveFileInput{
					Provider:        providerName,
					Data:            data,
					Path:            path,
					DefaultFileName: fileName,
				})
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(os.Stderr, fmtDataWrittenTo, out)
			}

			if c.Bool("stdout") {
				if err = output2.PrettyPrintJSON(data); err != nil {
					return fmt.Errorf("error printing data to stdout: %w", err)
				}
			}

			return nil
		},
	}
}
