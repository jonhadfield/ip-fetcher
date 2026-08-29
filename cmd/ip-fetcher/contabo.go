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
		providerName  = "contabo"
		fileName      = "prefixes.txt"
		fileNameLines = "contabo-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Contabo prefixes",
		Usage:     "Contabo",
		UsageText: "ip-fetcher contabo {--stdout | --Path FILE} [--lines]",
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

			data, err := contaboData(c, &h)
			if err != nil {
				return err
			}

			defaultName := fileName
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

// contaboData returns the newline separated prefixes when --lines is set, and
// the indented json document otherwise.
func contaboData(c *cli.Context, h *contabo.Contabo) ([]byte, error) {
	if c.Bool(formatLines) {
		doc, err := h.Fetch()
		if err != nil {
			return nil, err
		}

		return docToLines(doc)
	}

	raw, _, _, err := h.FetchData()
	if err != nil {
		return nil, err
	}

	var asnIPs contabo.Doc
	if err = json.Unmarshal(raw, &asnIPs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Contabo Data: %w", err)
	}

	data, err := json.MarshalIndent(asnIPs, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Contabo Data: %w", err)
	}

	return data, nil
}
