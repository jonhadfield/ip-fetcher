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
)

func gcpCmd() *cli.Command {
	return &cli.Command{
		Name:      providerNameGCP,
		HelpName:  "- fetch GCP prefixes",
		Usage:     "Google Cloud Platform",
		UsageText: "ip-fetcher gcp {--stdout | --Path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "Path",
				Usage: "where to save the file", Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: "write to stdout", Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "json, yaml, lines, csv", Value: "json", Aliases: []string{"f"},
			},
		},
		Action: func(c *cli.Context) error {
			path := strings.TrimSpace(c.String("Path"))
			if path == "" && !c.Bool("stdout") {
				_ = cli.ShowSubcommandHelp(c)
				fmt.Println("\nerror: must specify at least one of stdout and Path")
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
			// get Data if json output is requested
			if doc, err = a.Fetch(); err != nil {
				return err
			}

			return output(doc, c.String("format"), c.Bool("stdout"), c.String("Path"))
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
		if data = csv(doc); err != nil {
			return err
		}
	case "lines":
		if data, err = lines(doc); err != nil {
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
		fmt.Printf("%s\n", data)
	}

	if path != "" {
		var out string
		if out, err = SaveFile(SaveFileInput{
			Provider:        providerNameGCP,
			Data:            data,
			Path:            path,
			DefaultFileName: fileNameOutputGCP,
		}); err != nil {
			return err
		}

		if _, err = os.Stderr.WriteString(fmt.Sprintf("Data written to %s\n", out)); err != nil {
			return err
		}
	}

	return err
}

func lines(in gcp.Doc) ([]byte, error) {
	sl := strings.Builder{}
	for x := range in.IPv4Prefixes {
		sl.WriteString(fmt.Sprintf("%s\n", in.IPv4Prefixes[x].IPv4Prefix.String()))
	}

	for x := range in.IPv6Prefixes {
		sl.WriteString(fmt.Sprintf("%s\n", in.IPv6Prefixes[x].IPv6Prefix.String()))
	}

	return []byte(sl.String()), nil
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
