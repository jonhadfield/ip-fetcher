package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/alibaba"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func alibabaCmd() *cli.Command {
	const (
		providerName = "alibaba"
		fileName     = "prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Alibaba prefixes",
		Usage:     "Alibaba",
		UsageText: "ip-fetcher alibaba {--stdout | --Path FILE}",
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

			h := alibaba.New()

			if os.Getenv("IP_FETCHER_MOCK_ALIBABA") == "true" {
				defer gock.Off()
				urlBase := alibaba.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/alibaba/testdata/prefixes.json")
				gock.InterceptClient(h.Client.HTTPClient)
			}

			data, _, _, err := h.FetchData()
			if err != nil {
				return err
			}

			var asnIPs alibaba.Doc
			if err = json.Unmarshal(data, &asnIPs); err != nil {
				return fmt.Errorf("failed to unmarshal Alibaba Data: %w", err)
			}

			asnPrefixes, err := json.MarshalIndent(asnIPs, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal Alibaba Data: %w", err)
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
				_, _ = os.Stderr.WriteString(fmt.Sprintf(fmtDataWrittenTo, out))
			}

			if c.Bool("stdout") {
				fmt.Printf("%s\n", asnPrefixes)
			}

			return nil
		},
	}
}
