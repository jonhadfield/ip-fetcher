package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/bunny"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/yaml.v3"
)

const (
	providerNameBunny   = "bunny"
	fileNameOutputBunny = "bunny.json"
	fileNameLinesBunny  = "bunny-prefixes.txt"
)

var bunnyFormats = []string{formatJSON, formatYAML, formatLines, formatCSV}

func bunnyCmd() *cli.Command {
	return &cli.Command{
		Name:      providerNameBunny,
		HelpName:  "- fetch Bunny.net prefixes",
		Usage:     "Bunny.net",
		UsageText: "ip-fetcher bunny {--stdout | --Path FILE} [--lines]",
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
				Usage: strings.Join(bunnyFormats, ", "), Value: formatJSON, Aliases: []string{"f"},
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

			a := bunny.New()

			if isEnvEnabled("IP_FETCHER_MOCK_BUNNY") {
				defer gock.Off()

				v4, _ := url.Parse(bunny.IPv4URL)
				gock.New(fmt.Sprintf("%s://%s", v4.Scheme, v4.Host)).
					Get(v4.Path).
					Reply(http.StatusOK).
					File("../../providers/bunny/testdata/edgeserverlist.json")

				v6, _ := url.Parse(bunny.IPv6URL)
				gock.New(fmt.Sprintf("%s://%s", v6.Scheme, v6.Host)).
					Get(v6.Path).
					Reply(http.StatusOK).
					File("../../providers/bunny/testdata/edgeserverlist_ipv6.json")

				gock.InterceptClient(a.Client.HTTPClient)
			}

			var doc bunny.Doc
			if doc, err = a.Fetch(); err != nil {
				return err
			}

			format := c.String(flagFormat)
			if c.Bool(formatLines) {
				format = formatLines
			}

			return bunnyOutput(doc, format, stdout, path)
		},
	}
}

func bunnyOutput(doc bunny.Doc, format string, stdout bool, path string) error {
	if !slices.Contains(bunnyFormats, format) {
		return fmt.Errorf("invalid format: %s\n       choose from: %s",
			format, strings.Join(bunnyFormats, ", "))
	}

	var (
		data []byte
		err  error
	)

	switch format {
	case formatCSV:
		data = bunnyCsv(doc)
	case formatLines:
		data = bunnyLines(doc)
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

	defaultName := fileNameOutputBunny
	if format == formatLines {
		defaultName = fileNameLinesBunny
	}

	return writeOutputs(path, false, SaveFileInput{
		Provider:        providerNameBunny,
		DefaultFileName: defaultName,
		Data:            data,
	})
}

func bunnyLines(in bunny.Doc) []byte {
	sl := strings.Builder{}
	for x := range in.IPv4Prefixes {
		fmt.Fprintf(&sl, "%s\n", in.IPv4Prefixes[x].String())
	}

	for x := range in.IPv6Prefixes {
		fmt.Fprintf(&sl, "%s\n", in.IPv6Prefixes[x].String())
	}

	return []byte(sl.String())
}

func bunnyCsv(in bunny.Doc) []byte {
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
