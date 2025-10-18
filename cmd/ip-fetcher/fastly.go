package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/fastly"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/yaml.v3"
)

const (
	providerNameFastly   = "fastly"
	fileNameOutputFastly = "fastly.json"
	fileNameLines        = "fastly-prefixes.txt"
)

var fastlyFormats = []string{"json", "yaml", formatLines, "csv"}

func fastlyCmd() *cli.Command {
	return &cli.Command{
		Name:      providerNameFastly,
		HelpName:  "- fetch Fastly prefixes",
		Usage:     "Fastly",
		UsageText: "ip-fetcher fastly {--stdout | --Path FILE} [--lines]",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "Path",
				Usage: usageWhereToSaveFile, Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: usageWriteToStdout, Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: strings.Join(fastlyFormats, ", "), Value: "json", Aliases: []string{"f"},
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

			a := fastly.New()

			if isEnvEnabled("IP_FETCHER_MOCK_FASTLY") {
				defer gock.Off()
				urlBase := fastly.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/fastly/testdata/fastly.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			var doc fastly.Doc
			if doc, err = a.Fetch(); err != nil {
				return err
			}

			format := c.String("format")
			if c.Bool(formatLines) {
				format = formatLines
			}

			return fastlyOutput(doc, format, stdout, path)
		},
	}
}

func fastlyOutput(doc fastly.Doc, format string, stdout bool, path string) error {
	if !slices.Contains(fastlyFormats, format) {
		return fmt.Errorf("invalid format: %s\n       choose from: %s",
			format, strings.Join(fastlyFormats, ", "))
	}

	var (
		data []byte
		err  error
	)

	switch format {
	case "csv":
		data = fastlyCsv(doc)
	case formatLines:
		data = fastlyLines(doc)
	case "yaml":
		if data, err = yaml.Marshal(doc); err != nil {
			return err
		}
	case "json":
		if data, err = json.MarshalIndent(doc, "", " "); err != nil {
			return err
		}
	}

	if stdout {
		fmt.Printf("%s\n\n", data)
	}

	defaultName := fileNameOutputFastly
	if format == formatLines {
		defaultName = fileNameLines
	}

	return writeOutputs(path, false, SaveFileInput{
		Provider:        providerNameFastly,
		DefaultFileName: defaultName,
		Data:            data,
	})
}

func fastlyLines(in fastly.Doc) []byte {
	sl := strings.Builder{}
	for x := range in.IPv4Prefixes {
		sl.WriteString(fmt.Sprintf("%s\n", in.IPv4Prefixes[x].String()))
	}

	for x := range in.IPv6Prefixes {
		sl.WriteString(fmt.Sprintf("%s\n", in.IPv6Prefixes[x].String()))
	}

	return []byte(sl.String())
}

func fastlyCsv(in fastly.Doc) []byte {
	sl := strings.Builder{}
	for x := range in.IPv4Prefixes {
		sl.WriteString(fmt.Sprintf("\"%s\"", in.IPv4Prefixes[x].String()))
		// output comma if not last line and there are ipv6 prefixes
		if x != len(in.IPv4Prefixes)-1 && len(in.IPv6Prefixes) > 0 {
			sl.WriteString(",\n")
		}
	}

	for x := range in.IPv6Prefixes {
		sl.WriteString(fmt.Sprintf("\"%s\"", in.IPv6Prefixes[x].String()))

		if x != len(in.IPv6Prefixes)-1 {
			sl.WriteString(",\n")
		}
	}

	sl.WriteString("\n")

	return []byte(sl.String())
}
