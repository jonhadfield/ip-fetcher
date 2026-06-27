package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/imperva"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/yaml.v3"
)

const (
	providerNameImperva   = "imperva"
	fileNameOutputImperva = "imperva.json"
	fileNameLinesImperva  = "imperva-prefixes.txt"
)

var impervaFormats = []string{formatJSON, formatYAML, formatLines}

func impervaCmd() *cli.Command {
	return &cli.Command{
		Name:      providerNameImperva,
		HelpName:  "- fetch Imperva prefixes",
		Usage:     "Imperva",
		UsageText: "ip-fetcher imperva {--stdout | --Path FILE} [--lines]",
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
				Usage: strings.Join(impervaFormats, ", "), Value: formatJSON, Aliases: []string{"f"},
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

			a := imperva.New()

			if isEnvEnabled("IP_FETCHER_MOCK_IMPERVA") {
				defer gock.Off()
				u, _ := url.Parse(imperva.DownloadURL)
				gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
					Post(u.Path).
					Reply(http.StatusOK).
					File("../../providers/imperva/testdata/ips.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			var doc imperva.Doc
			if doc, err = a.Fetch(); err != nil {
				return err
			}

			format := c.String(flagFormat)
			if c.Bool(formatLines) {
				format = formatLines
			}

			return impervaOutput(doc, format, stdout, path)
		},
	}
}

func impervaOutput(doc imperva.Doc, format string, stdout bool, path string) error {
	if !slices.Contains(impervaFormats, format) {
		return fmt.Errorf("invalid format: %s\n       choose from: %s",
			format, strings.Join(impervaFormats, ", "))
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

	defaultName := fileNameOutputImperva
	if format == formatLines {
		defaultName = fileNameLinesImperva
	}

	return writeOutputs(path, false, SaveFileInput{
		Provider:        providerNameImperva,
		DefaultFileName: defaultName,
		Data:            data,
	})
}
