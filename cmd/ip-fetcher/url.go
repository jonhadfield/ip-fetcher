package main

import (
	"errors"
	"net/http"
	"net/url"
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
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			h := _url.New()
			var requests []_url.Request
			for _, u := range urlList {
				parsedURL, parseErr := url.Parse(u)
				if parseErr != nil {
					continue
				}

				requests = append(requests, _url.Request{
					URL:    parsedURL,
					Method: "GET",
				})
			}

			if isEnvEnabled("IP_FETCHER_MOCK_URL") {
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

			joined := strings.Join(prefixes, "\n")

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        providerName,
				DefaultFileName: fileName,
				Data:            []byte(joined),
			})
		},
	}
}
