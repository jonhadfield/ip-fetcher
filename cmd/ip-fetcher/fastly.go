package main

import (
	"encoding/json"
	"fmt"
	"github.com/jonhadfield/ip-fetcher/providers/fastly"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"slices"
	"strings"
)

const (
	providerNameFastly   = "fastly"
	fileNameOutputFastly = "fastly.json"
)

var fastlyFormats = []string{"json", "yaml", "lines", "csv"}

func fastlyCmd() *cli.Command {
	return &cli.Command{
		Name:      providerNameFastly,
		HelpName:  "- fetch Fastly prefixes",
		Usage:     "Fastly",
		UsageText: "ip-fetcher fastly {--stdout | --path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "where to save the file", Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: "write to stdout", Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: strings.Join(fastlyFormats, ", "), Value: "json", Aliases: []string{"f"},
			},
		},
		Action: func(c *cli.Context) error {
			path := strings.TrimSpace(c.String("path"))
			if path == "" && !c.Bool("stdout") {
				_ = cli.ShowSubcommandHelp(c)
				fmt.Println("\nerror: must specify at least one of stdout and path")
				os.Exit(1)
			}

			a := fastly.New()

			if os.Getenv("IP_FETCHER_MOCK_Fastly") == "true" {
				defer gock.Off()
				urlBase := fastly.DefaultUrl
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(200).
					File("../../providers/fastly/testdata/fastly.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			var doc fastly.Doc
			var err error
			// get data if json output is requested
			if doc, err = a.Fetch(); err != nil {
				return err
			}

			return fastlyOutput(doc, c.String("format"), c.Bool("stdout"), c.String("path"))
		},
	}
}

func fastlyOutput(doc fastly.Doc, format string, stdout bool, path string) (err error) {
	if !slices.Contains(fastlyFormats, format) {
		return fmt.Errorf("invalid format: %s\n       choose from: %s",
			format, strings.Join(fastlyFormats, ", "))
	}

	var data []byte

	switch format {
	case "csv":
		if data, err = fastlyCsv(doc); err != nil {
			return err
		}
	case "lines":
		if data, err = fastlyLines(doc); err != nil {
			return err
		}
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

	if path != "" {
		var out string
		if out, err = saveFile(saveFileInput{
			provider:        providerNameFastly,
			data:            data,
			path:            path,
			defaultFileName: fileNameOutputFastly,
		}); err != nil {
			return err
		}

		if _, err = os.Stderr.WriteString(fmt.Sprintf("data written to %s\n", out)); err != nil {
			return err
		}
	}

	return err
}

func fastlyLines(in fastly.Doc) ([]byte, error) {
	sl := strings.Builder{}
	for x := range in.IPv4Prefixes {
		sl.WriteString(fmt.Sprintf("%s\n", in.IPv4Prefixes[x].String()))
	}
	for x := range in.IPv6Prefixes {
		sl.WriteString(fmt.Sprintf("%s\n", in.IPv6Prefixes[x].String()))
	}

	return []byte(sl.String()), nil
}

func fastlyCsv(in fastly.Doc) ([]byte, error) {
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

	return []byte(sl.String()), nil
}
