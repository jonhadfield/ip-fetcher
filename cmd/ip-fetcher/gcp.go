package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jonhadfield/ip-fetcher/providers/gcp"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

const (
	providerNameGCP   = "gcp"
	fileNameOutputGCP = "cloud.json"
	fileNameLinesGCP  = "gcp-prefixes.txt"
)

func gcpCmd() *cli.Command {
	return &cli.Command{
		Name:      providerNameGCP,
		HelpName:  "- fetch GCP prefixes",
		Usage:     "Google Cloud Platform",
		UsageText: "ip-fetcher gcp {--stdout | --Path FILE} [--lines]",
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
				Usage: "json, yaml, lines, csv", Value: "json", Aliases: []string{"f"},
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

			a := gcp.New()

			if isEnvEnabled("IP_FETCHER_MOCK_GCP") {
				defer gock.Off()
				urlBase := gcp.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/gcp/testdata/cloud.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			var doc gcp.Doc
			// fetch document
			if doc, err = a.Fetch(); err != nil {
				return err
			}

			format := c.String("format")
			if c.Bool(formatLines) {
				format = formatLines
			}

			return output(doc, format, stdout, path)
		},
	}
}

func output(doc gcp.Doc, format string, stdout bool, path string) error {
	var (
		data []byte
		err  error
	)

	switch format {
	case "csv":
		data = csv(doc)
	case formatLines:
		data = lines(doc)
	case "yaml":
		jsonPayload, marshalErr := json.Marshal(doc)
		if marshalErr != nil {
			return marshalErr
		}

		var intermediate map[string]any
		if err = json.Unmarshal(jsonPayload, &intermediate); err != nil {
			return err
		}

		if data, err = yaml.Marshal(intermediate); err != nil {
			return err
		}
	case "json":
		if data, err = json.MarshalIndent(doc, "", " "); err != nil {
			return err
		}
	}

	defaultName := fileNameOutputGCP
	if format == formatLines {
		defaultName = fileNameLinesGCP
	}

	return writeOutputs(path, stdout, SaveFileInput{
		Provider:        providerNameGCP,
		DefaultFileName: defaultName,
		Data:            data,
	})
}

func lines(in gcp.Doc) []byte {
	sl := strings.Builder{}
	for x := range in.IPv4Prefixes {
		sl.WriteString(fmt.Sprintf("%s\n", in.IPv4Prefixes[x].IPv4Prefix.String()))
	}

	for x := range in.IPv6Prefixes {
		sl.WriteString(fmt.Sprintf("%s\n", in.IPv6Prefixes[x].IPv6Prefix.String()))
	}

	return []byte(sl.String())
}

func csv(in gcp.Doc) []byte {
	sl := strings.Builder{}
	for x := range in.IPv4Prefixes {
		sl.WriteString(fmt.Sprintf("\"%s\"", in.IPv4Prefixes[x].IPv4Prefix.String()))
		// output comma if not last line and there are ipv6 prefixes
		if x != len(in.IPv4Prefixes)-1 && len(in.IPv6Prefixes) > 0 {
			sl.WriteString(",\n")
		}
	}

	for x := range in.IPv6Prefixes {
		sl.WriteString(fmt.Sprintf("\"%s\"", in.IPv6Prefixes[x].IPv6Prefix.String()))
		if x != len(in.IPv6Prefixes)-1 {
			sl.WriteString(",\n")
		}
	}

	return []byte(sl.String())
}
