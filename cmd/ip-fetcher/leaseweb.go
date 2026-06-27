package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/leaseweb"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func leasewebCmd() *cli.Command {
	const (
		providerName = "leaseweb"
		fileName     = "prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Leaseweb prefixes",
		Usage:     "Leaseweb",
		UsageText: "ip-fetcher leaseweb {--stdout | --Path FILE}",
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

			h := leaseweb.New()

			if isEnvEnabled("IP_FETCHER_MOCK_LEASEWEB") {
				defer gock.Off()
				for _, asn := range leaseweb.ASNs {
					urlBase := fmt.Sprintf(leaseweb.DownloadURL, asn)
					u, _ := url.Parse(urlBase)
					gock.New(urlBase).
						Get(u.Path).
						Reply(http.StatusOK).
						File("../../providers/leaseweb/testdata/prefixes.json")
				}
				gock.InterceptClient(h.Client.HTTPClient)
			}

			data, _, _, err := h.FetchData()
			if err != nil {
				return err
			}

			var asnIPs leaseweb.Doc
			if err = json.Unmarshal(data, &asnIPs); err != nil {
				return fmt.Errorf("failed to unmarshal Leaseweb Data: %w", err)
			}

			asnPrefixes, err := json.MarshalIndent(asnIPs, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal Leaseweb Data: %w", err)
			}

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        providerName,
				DefaultFileName: fileName,
				Data:            asnPrefixes,
			})
		},
	}
}
