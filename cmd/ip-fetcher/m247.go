package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/m247"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func m247Cmd() *cli.Command {
	const (
		providerName  = "m247"
		fileName      = "prefixes.txt"
		fileNameLines = "m247-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch M247 prefixes",
		Usage:     "M247",
		UsageText: "ip-fetcher m247 {--stdout | --Path FILE} [--lines]",
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

			h := m247.New()

			if isEnvEnabled("IP_FETCHER_MOCK_M247") {
				defer gock.Off()
				urlBase := fmt.Sprintf(m247.DownloadURL, "16247")
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/m247/testdata/prefixes.json")
				gock.InterceptClient(h.Client.HTTPClient)
			}

			data, err := m247Data(c, &h)
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

// m247Data returns the newline separated prefixes when --lines is set, and the
// indented json document otherwise.
func m247Data(c *cli.Context, h *m247.M247) ([]byte, error) {
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

	var asnIPs m247.Doc
	if err = json.Unmarshal(raw, &asnIPs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal M247 Data: %w", err)
	}

	data, err := json.MarshalIndent(asnIPs, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal M247 Data: %w", err)
	}

	return data, nil
}
