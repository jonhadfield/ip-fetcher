package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/github"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func githubCmd() *cli.Command {
	const (
		providerName = "github"
		fileName     = "prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch GitHub prefixes",
		Usage:     "GitHub",
		UsageText: "ip-fetcher github {--stdout | --Path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)
			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "Path",
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
			path := strings.TrimSpace(c.String("Path"))
			if path == "" && !c.Bool("stdout") {
				_ = cli.ShowSubcommandHelp(c)
				fmt.Println("\nerror: must specify at least one of stdout and Path")
				os.Exit(1)
			}

			gh := github.New()

			if os.Getenv("IP_FETCHER_MOCK_GITHUB") == "true" {
				defer gock.Off()
				urlBase := github.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/github/testdata/meta.json")
				gock.InterceptClient(gh.Client.HTTPClient)
			}

			prefixes, err := gh.Fetch()
			if err != nil {
				return err
			}

			var lines []string
			for _, p := range prefixes {
				lines = append(lines, p.String())
			}
			data := []byte(strings.Join(lines, "\n"))

			if path != "" {
				out, err := SaveFile(SaveFileInput{
					Provider:        providerName,
					Data:            data,
					Path:            path,
					DefaultFileName: fileName,
				})
				if err != nil {
					return err
				}
				_, _ = os.Stderr.WriteString(fmt.Sprintf("Data written to %s\n", out))
			}

			if c.Bool("stdout") {
				fmt.Printf("%s\n", data)
			}

			return nil
		},
	}
}
