package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/scaleway"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func scalewayCmd() *cli.Command {
	const (
		providerName  = "scaleway"
		fileName      = "prefixes.txt"
		fileNameLines = "scaleway-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Scaleway prefixes",
		Usage:     "Scaleway",
		UsageText: "ip-fetcher scaleway {--stdout | --Path FILE} [--lines]",
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

			h := scaleway.New()

			if isEnvEnabled("IP_FETCHER_MOCK_SCALEWAY") {
				defer gock.Off()
				urlBase := fmt.Sprintf(scaleway.DownloadURL, "12876")
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/scaleway/testdata/prefixes.json")
				gock.InterceptClient(h.Client.HTTPClient)
			}

			data, err := scalewayData(c, &h)
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

// scalewayData returns the newline separated prefixes when --lines is set, and the
// indented json document otherwise.
func scalewayData(c *cli.Context, h *scaleway.Scaleway) ([]byte, error) {
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

	var asnIPs scaleway.Doc
	if err = json.Unmarshal(raw, &asnIPs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Scaleway Data: %w", err)
	}

	data, err := json.MarshalIndent(asnIPs, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Scaleway Data: %w", err)
	}

	return data, nil
}
