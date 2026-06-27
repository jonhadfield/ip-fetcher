package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/atlassian"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/yaml.v3"
)

const (
	providerNameAtlassian   = "atlassian"
	fileNameOutputAtlassian = "atlassian.json"
	fileNameLinesAtlassian  = "atlassian-prefixes.txt"
)

var atlassianFormats = []string{formatJSON, formatYAML, formatLines}

func atlassianCmd() *cli.Command {
	return &cli.Command{
		Name:      providerNameAtlassian,
		HelpName:  "- fetch Atlassian prefixes",
		Usage:     "Atlassian",
		UsageText: "ip-fetcher atlassian {--stdout | --Path FILE} [--lines]",
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
				Usage: strings.Join(atlassianFormats, ", "), Value: formatJSON, Aliases: []string{"f"},
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

			a := atlassian.New()

			if isEnvEnabled("IP_FETCHER_MOCK_ATLASSIAN") {
				defer gock.Off()
				u, _ := url.Parse(atlassian.DownloadURL)
				gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/atlassian/testdata/ip-ranges.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			var doc atlassian.Doc
			if doc, err = a.Fetch(); err != nil {
				return err
			}

			format := c.String(flagFormat)
			if c.Bool(formatLines) {
				format = formatLines
			}

			return atlassianOutput(doc, format, stdout, path)
		},
	}
}

func atlassianOutput(doc atlassian.Doc, format string, stdout bool, path string) error {
	if !slices.Contains(atlassianFormats, format) {
		return fmt.Errorf("invalid format: %s\n       choose from: %s",
			format, strings.Join(atlassianFormats, ", "))
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

	defaultName := fileNameOutputAtlassian
	if format == formatLines {
		defaultName = fileNameLinesAtlassian
	}

	return writeOutputs(path, false, SaveFileInput{
		Provider:        providerNameAtlassian,
		DefaultFileName: defaultName,
		Data:            data,
	})
}
