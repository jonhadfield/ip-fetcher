package main

import (
	"fmt"
	_url "github.com/jonhadfield/ip-fetcher/providers/url"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
	"net/url"
	"os"
	"strings"
)

func urlCmd() *cli.Command {
	const (
		providerName = "url"
		fileName     = "ips.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch prefixes from URLs",
		Usage:     "URL",
		UsageText: "ip-fetcher url {--stdout | --path FILE} URL [URL...]",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "where to save the file", Aliases: []string{"p"}, TakesFile: true,
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: "write to stdout", Aliases: []string{"s"},
			},
		},
		Action: func(c *cli.Context) error {
			urls := c.Args().Slice()
			path := strings.TrimSpace(c.String("path"))
			if path == "" && !c.Bool("stdout") {
				_ = cli.ShowSubcommandHelp(c)

				fmt.Println("\nerror: must specify at least one of stdout and path")

				os.Exit(1)
			}

			h := _url.New()
			h.Add(urls)

			if os.Getenv("IP_FETCHER_MOCK_URL") == "true" {
				defer gock.Off()
				urlBase := "https://www.example.com/files/ips.txt"
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(200).
					File("../../providers/url/testdata/ip-file-1.txt")
				gock.InterceptClient(h.Client.HTTPClient)
			}

			prefixes, err := h.FetchPrefixesAsText()
			if err != nil {
				return err
			}

			if len(prefixes) == 0 {
				return fmt.Errorf("no prefixes found")
			}

			if path != "" {
				var out string

				if out, err = saveFile(saveFileInput{
					provider:        providerName,
					data:            []byte(strings.Join(prefixes, "\n")),
					path:            path,
					defaultFileName: fileName,
				}); err != nil {
					return err
				}

				_, _ = os.Stderr.WriteString(fmt.Sprintf("data written to %s\n", out))
			}

			if c.Bool("stdout") {
				fmt.Printf("%s\n", strings.Join(prefixes, "\n"))
			}

			return nil
		},
	}
}