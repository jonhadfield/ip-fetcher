package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"strings"

	output2 "github.com/jonhadfield/ip-fetcher/internal/output"

	"github.com/jonhadfield/ip-fetcher/providers/zscaler"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func zscalerCmd() *cli.Command {
	const (
		providerName  = "zscaler"
		fileName      = "prefixes.txt"
		fileNameLines = "zscaler-prefixes.txt"
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

			if c.Bool(formatLines) {
				var raw []byte
				if raw, _, _, err = z.FetchData(); err != nil {
					return err
				}

				var lines []byte
				if lines, err = docToLines(zscalerPrefixes(raw)); err != nil {
					return err
				}

				return writeOutputs(path, stdout, SaveFileInput{
					Provider:        providerName,
					DefaultFileName: fileNameLines,
					Data:            lines,
				})
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

// zscalerPrefixes pulls the range values out of the zscaler document.
//
// zscaler.Doc carries every range as a string rather than a netip.Prefix, and
// declares one field per city, so a reflective walk finds no prefixes and
// silently drops any city the struct does not name. Walking the decoded json
// avoids both problems.
func zscalerPrefixes(raw []byte) []netip.Prefix {
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil
	}

	var prefixes []netip.Prefix

	collectZscalerRanges(doc, &prefixes)

	return prefixes
}

// collectZscalerRanges walks the decoded document, gathering every "range".
func collectZscalerRanges(v any, out *[]netip.Prefix) {
	switch typed := v.(type) {
	case map[string]any:
		for key, value := range typed {
			if key == "range" {
				addZscalerRange(value, out)

				continue
			}

			collectZscalerRanges(value, out)
		}
	case []any:
		for _, item := range typed {
			collectZscalerRanges(item, out)
		}
	}
}

// addZscalerRange appends value if it parses as a prefix.
func addZscalerRange(value any, out *[]netip.Prefix) {
	s, ok := value.(string)
	if !ok {
		return
	}

	prefix, err := netip.ParsePrefix(strings.TrimSpace(s))
	if err != nil {
		return
	}

	*out = append(*out, prefix)
}
