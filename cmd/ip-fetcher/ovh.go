package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/ovh"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func ovhCmd() *cli.Command {
	const (
		providerName = "ovh"
		fileName     = "prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch OVH prefixes",
		Usage:     "OVHcloud",
		UsageText: "ip-fetcher ovh {--stdout | --path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)
			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Usage:   "where to save the file",
				Aliases: []string{"p"}, TakesFile: true,
			},
			&cli.BoolFlag{
				Name:    "stdout",
				Usage:   "write to stdout",
				Aliases: []string{"s"},
			},
		},
		Action: func(c *cli.Context) error {
			path := strings.TrimSpace(c.String("path"))
			if path == "" && !c.Bool("stdout") {
				_ = cli.ShowSubcommandHelp(c)
				fmt.Println("\nerror: must specify at least one of stdout and path")
				os.Exit(1)
			}

			o := ovh.New()

			if os.Getenv("IP_FETCHER_MOCK_OVH") == "true" {
				defer gock.Off()
				urlBase := ovh.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(200).
					File("../../providers/ovh/testdata/prefixes.txt")
				gock.InterceptClient(o.Client.HTTPClient)
			}

			prefixes, err := o.Fetch()
			if err != nil {
				return err
			}

			var lines []string
			for _, p := range prefixes {
				lines = append(lines, p.String())
			}
			data := []byte(strings.Join(lines, "\n"))

			if path != "" {
				out, err := saveFile(saveFileInput{
					provider:        providerName,
					data:            data,
					path:            path,
					defaultFileName: fileName,
				})
				if err != nil {
					return err
				}
				_, _ = os.Stderr.WriteString(fmt.Sprintf("data written to %s\n", out))
			}

			if c.Bool("stdout") {
				fmt.Printf("%s\n", data)
			}

			return nil
		},
	}
}
