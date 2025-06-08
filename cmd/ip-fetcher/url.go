package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	_url "github.com/jonhadfield/ip-fetcher/providers/url"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func urlCmd() *cli.Command {
	const (
		providerName = "url"
		fileName     = "ips.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch prefixes from URLs",
		Usage:     "Read prefixes from a web URL",
		UsageText: "ip-fetcher url {--stdout | --Path FILE} URL [URL...]",
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
		},
		Action: func(c *cli.Context) error {
			urlList := c.Args().Slice()
			path := strings.TrimSpace(c.String("Path"))
			if path == "" && !c.Bool("stdout") {
				_ = cli.ShowSubcommandHelp(c)

				fmt.Println("\n" + errStdoutOrPathRequired)

				os.Exit(1)
			}

			h := _url.New()
			var requests []_url.Request
			for _, u := range urlList {
				var pURL *url.URL
				var err error
				if pURL, err = url.Parse(u); err != nil {
					continue
				}

				requests = append(requests, _url.Request{
					URL:    pURL,
					Method: "GET",
				})
			}

			if os.Getenv("IP_FETCHER_MOCK_URL") == "true" {
				defer gock.Off()
				urlBase := "https://www.example.com/files/ips.txt"
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/url/testdata/ip-file-1.txt")
				gock.InterceptClient(h.HTTPClient.HTTPClient)
			}

			prefixes, err := h.FetchPrefixesAsText(requests)
			if err != nil {
				return err
			}

			if len(prefixes) == 0 {
				return errors.New("no prefixes found")
			}

			if path != "" {
				var out string

				if out, err = SaveFile(SaveFileInput{
					Provider:        providerName,
					Data:            []byte(strings.Join(prefixes, "\n")),
					Path:            path,
					DefaultFileName: fileName,
				}); err != nil {
					return err
				}

				_, _ = fmt.Fprintf(os.Stderr, fmtDataWrittenTo, out)
			}

			if c.Bool("stdout") {
				fmt.Printf("%s\n", strings.Join(prefixes, "\n"))
			}

			return nil
		},
	}
}
