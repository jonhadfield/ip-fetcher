package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/contabo"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func contaboCmd() *cli.Command {
	const (
		providerName = "contabo"
		fileName     = "prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Contabo prefixes",
		Usage:     "Contabo",
		UsageText: "ip-fetcher contabo {--stdout | --Path FILE}",
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

			h := contabo.New()

			if isEnvEnabled("IP_FETCHER_MOCK_CONTABO") {
				defer gock.Off()
				urlBase := fmt.Sprintf(contabo.DownloadURL, contabo.ASNs[0])
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/contabo/testdata/prefixes.json")
				gock.InterceptClient(h.Client.HTTPClient)
			}

			data, _, _, err := h.FetchData()
			if err != nil {
				return err
			}

			var asnIPs contabo.Doc
			if err = json.Unmarshal(data, &asnIPs); err != nil {
				return fmt.Errorf("failed to unmarshal Contabo Data: %w", err)
			}

			asnPrefixes, err := json.MarshalIndent(asnIPs, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal Contabo Data: %w", err)
			}

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        providerName,
				DefaultFileName: fileName,
				Data:            asnPrefixes,
			})
		},
	}
}
