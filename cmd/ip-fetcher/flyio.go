package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/flyio"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func flyioCmd() *cli.Command {
	const (
		providerName  = "flyio"
		fileName      = "prefixes.txt"
		fileNameLines = "flyio-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Fly.io prefixes",
		Usage:     "Fly.io",
		UsageText: "ip-fetcher flyio {--stdout | --Path FILE} [--lines]",
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

			h := flyio.New()

			if isEnvEnabled("IP_FETCHER_MOCK_FLYIO") {
				defer gock.Off()
				urlBase := fmt.Sprintf(flyio.DownloadURL, flyio.ASNs[0])
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/flyio/testdata/prefixes.json")
				gock.InterceptClient(h.Client.HTTPClient)
			}

			data, err := flyioData(c, &h)
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

// flyioData returns the newline separated prefixes when --lines is set, and
// the indented json document otherwise.
func flyioData(c *cli.Context, h *flyio.Flyio) ([]byte, error) {
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

	var asnIPs flyio.Doc
	if err = json.Unmarshal(raw, &asnIPs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Fly.io Data: %w", err)
	}

	data, err := json.MarshalIndent(asnIPs, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Fly.io Data: %w", err)
	}

	return data, nil
}
