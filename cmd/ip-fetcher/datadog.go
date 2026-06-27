package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/datadog"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/yaml.v3"
)

const (
	providerNameDatadog   = "datadog"
	fileNameOutputDatadog = "datadog.json"
	fileNameLinesDatadog  = "datadog-prefixes.txt"
)

var datadogFormats = []string{formatJSON, formatYAML, formatLines}

func datadogCmd() *cli.Command {
	return &cli.Command{
		Name:      providerNameDatadog,
		HelpName:  "- fetch Datadog prefixes",
		Usage:     "Datadog",
		UsageText: "ip-fetcher datadog {--stdout | --Path FILE} [--lines]",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagPath,
				Usage: usageWhereToSaveFile, Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  flagStdout,
				Usage: usageWriteToStdout, Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:  flagFormat,
				Usage: strings.Join(datadogFormats, ", "), Value: formatJSON, Aliases: []string{"f"},
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

			a := datadog.New()

			if isEnvEnabled("IP_FETCHER_MOCK_DATADOG") {
				defer gock.Off()
				u, _ := url.Parse(datadog.DownloadURL)
				gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/datadog/testdata/ip-ranges.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			var doc datadog.Doc
			if doc, err = a.Fetch(); err != nil {
				return err
			}

			format := c.String(flagFormat)
			if c.Bool(formatLines) {
				format = formatLines
			}

			return datadogOutput(doc, format, stdout, path)
		},
	}
}

func datadogOutput(doc datadog.Doc, format string, stdout bool, path string) error {
	if !slices.Contains(datadogFormats, format) {
		return fmt.Errorf("invalid format: %s\n       choose from: %s",
			format, strings.Join(datadogFormats, ", "))
	}

	var (
		data []byte
		err  error
	)

	switch format {
	case formatLines:
		data = prefixesToLines(doc.IPv4Prefixes, doc.IPv6Prefixes)
	case formatYAML:
		if data, err = yaml.Marshal(doc); err != nil {
			return err
		}
	case formatJSON:
		if data, err = json.MarshalIndent(doc, "", " "); err != nil {
			return err
		}
	}

	if stdout {
		fmt.Printf("%s\n\n", data)
	}

	defaultName := fileNameOutputDatadog
	if format == formatLines {
		defaultName = fileNameLinesDatadog
	}

	return writeOutputs(path, false, SaveFileInput{
		Provider:        providerNameDatadog,
		DefaultFileName: defaultName,
		Data:            data,
	})
}
