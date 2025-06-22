package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/digitalocean"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func digitaloceanCmd() *cli.Command {
	const (
		fileNameData  = "google.csv"
		fileNameLines = "digitalocean-prefixes.txt"
	)

	return &cli.Command{
		Name:      "digitalocean",
		HelpName:  "- fetch DigitalOcean prefixes",
		Usage:     "DigitalOcean",
		UsageText: "ip-fetcher digitalocean {--stdout | --Path FILE} [--lines]",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "Path",
				Usage: usageWhereToSaveFile, Aliases: []string{"p"}, TakesFile: true,
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: usageWriteToStdout, Aliases: []string{"s"},
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

			a := digitalocean.New()

			if os.Getenv("IP_FETCHER_MOCK_DIGITALOCEAN") == "true" {
				defer gock.Off()
				urlBase := digitalocean.DigitaloceanDownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/digitalocean/testdata/google.csv")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			var data []byte
			var err error
			if c.Bool("lines") {
				var doc digitalocean.Doc
				if doc, err = a.Fetch(); err != nil {
					return err
				}
				if data, err = docToLines(doc); err != nil {
					return err
				}
			} else {
				data, _, _, err = a.FetchData()
				if err != nil {
					return err
				}
			}

			if path != "" {
				var out string

				df := fileNameData
				if c.Bool("lines") {
					df = fileNameLines
				}
				if out, err = SaveFile(SaveFileInput{
					Provider:        "digitalocean",
					Data:            data,
					Path:            path,
					DefaultFileName: df,
				}); err != nil {
					return err
				}

				_, _ = fmt.Fprintf(os.Stderr, fmtDataWrittenTo, out)
			}

			if c.Bool("stdout") {
				fmt.Printf("%s\n", data)
			}

			return nil
		},
	}
}
