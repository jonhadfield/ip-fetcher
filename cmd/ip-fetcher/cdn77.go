package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/cdn77"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/yaml.v3"
)

const (
	providerNameCDN77   = "cdn77"
	fileNameOutputCDN77 = "cdn77.json"
	fileNameLinesCDN77  = "cdn77-prefixes.txt"
)

var cdn77Formats = []string{formatJSON, formatYAML, formatLines, formatCSV}

func cdn77Cmd() *cli.Command {
	return &cli.Command{
		Name:      providerNameCDN77,
		HelpName:  "- fetch CDN77 prefixes",
		Usage:     "CDN77",
		UsageText: "ip-fetcher cdn77 {--stdout | --Path FILE} [--lines]",
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
				Usage: strings.Join(cdn77Formats, ", "), Value: formatJSON, Aliases: []string{"f"},
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

			a := cdn77.New()

			if isEnvEnabled("IP_FETCHER_MOCK_CDN77") {
				defer gock.Off()
				urlBase := cdn77.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/cdn77/testdata/prefixes.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			var doc cdn77.Doc
			if doc, err = a.Fetch(); err != nil {
				return err
			}

			format := c.String(flagFormat)
			if c.Bool(formatLines) {
				format = formatLines
			}

			return cdn77Output(doc, format, stdout, path)
		},
	}
}

func cdn77Output(doc cdn77.Doc, format string, stdout bool, path string) error {
	if !slices.Contains(cdn77Formats, format) {
		return fmt.Errorf("invalid format: %s\n       choose from: %s",
			format, strings.Join(cdn77Formats, ", "))
	}

	var (
		data []byte
		err  error
	)

	switch format {
	case formatCSV:
		data = cdn77Csv(doc)
	case formatLines:
		data = cdn77Lines(doc)
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

	defaultName := fileNameOutputCDN77
	if format == formatLines {
		defaultName = fileNameLinesCDN77
	}

	return writeOutputs(path, false, SaveFileInput{
		Provider:        providerNameCDN77,
		DefaultFileName: defaultName,
		Data:            data,
	})
}

func cdn77Lines(in cdn77.Doc) []byte {
	sl := strings.Builder{}
	for x := range in.IPv4Prefixes {
		fmt.Fprintf(&sl, "%s\n", in.IPv4Prefixes[x].String())
	}

	for x := range in.IPv6Prefixes {
		fmt.Fprintf(&sl, "%s\n", in.IPv6Prefixes[x].String())
	}

	return []byte(sl.String())
}

func cdn77Csv(in cdn77.Doc) []byte {
	sl := strings.Builder{}
	for x := range in.IPv4Prefixes {
		fmt.Fprintf(&sl, "\"%s\"", in.IPv4Prefixes[x].String())
		if x != len(in.IPv4Prefixes)-1 && len(in.IPv6Prefixes) > 0 {
			sl.WriteString(",\n")
		}
	}

	for x := range in.IPv6Prefixes {
		fmt.Fprintf(&sl, "\"%s\"", in.IPv6Prefixes[x].String())

		if x != len(in.IPv6Prefixes)-1 {
			sl.WriteString(",\n")
		}
	}

	sl.WriteString("\n")

	return []byte(sl.String())
}
