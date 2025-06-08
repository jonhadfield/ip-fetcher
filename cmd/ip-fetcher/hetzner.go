package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/hetzner"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func hetznerCmd() *cli.Command {
	const (
		providerName = "hetzner"
		fileName     = "prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Hetzner prefixes",
		Usage:     "Hetzner",
		UsageText: "ip-fetcher hetzner {--stdout | --Path FILE}",
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

			h := hetzner.New()

			if os.Getenv("IP_FETCHER_MOCK_HETZNER") == "true" {
				defer gock.Off()
				urlBase := hetzner.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/hetzner/testdata/prefixes.json")
				gock.InterceptClient(h.Client.HTTPClient)
			}

			data, _, _, err := h.FetchData()
			if err != nil {
				return err
			}

			var asnIPs hetzner.Doc
			if err = json.Unmarshal(data, &asnIPs); err != nil {
				return fmt.Errorf("failed to unmarshal Hetzner Data: %w", err)
			}

			asnPrefixes, err := json.MarshalIndent(asnIPs, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal Hetzner Data: %w", err)
			}

			var out string
			if path != "" {
				out, err = SaveFile(SaveFileInput{
					Provider:        providerName,
					Data:            asnPrefixes,
					Path:            path,
					DefaultFileName: fileName,
				})
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(os.Stderr, fmtDataWrittenTo, out)
			}

			if c.Bool("stdout") {
				fmt.Printf("%s\n", asnPrefixes)
			}

			return nil
		},
	}
}
