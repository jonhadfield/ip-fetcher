package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
				Name:  "lines",
				Usage: usageLinesOutput,
			},
		},
		Action: func(c *cli.Context) error {
			path := strings.TrimSpace(c.String("Path"))
			if path == "" && !c.Bool("stdout") {
				_ = cli.ShowSubcommandHelp(c)
				fmt.Println("\n" + errStdoutOrPathRequired)
				os.Exit(1)
			}

			a := gcp.New()

			if os.Getenv("IP_FETCHER_MOCK_GCP") == "true" {
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
			var err error
			// fetch document
			if doc, err = a.Fetch(); err != nil {
				return err
			}

			format := c.String("format")
			if c.Bool("lines") {
				format = "lines"
			}

			return output(doc, format, c.Bool("stdout"), c.String("Path"))
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
	case "lines":
		data = lines(doc)
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
		fmt.Printf("%s\n", data)
	}

	if path != "" {
		var out string
		df := fileNameOutputGCP
		if format == "lines" {
			df = fileNameLinesGCP
		}
		if out, err = SaveFile(SaveFileInput{
			Provider:        providerNameGCP,
			Data:            data,
			Path:            path,
			DefaultFileName: df,
		}); err != nil {
			return err
		}

		if _, err = fmt.Fprintf(os.Stderr, fmtDataWrittenTo, out); err != nil {
			return err
		}
	}

	return err
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
