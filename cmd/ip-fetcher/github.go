package main

import (
	"net/http"
	"net/url"
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
				Usage:   usageWhereToSaveFile,
				Aliases: []string{"p"}, TakesFile: true,
			},
			&cli.BoolFlag{
				Name:    "stdout",
				Usage:   usageWriteToStdout,
				Aliases: []string{"s"},
			},
		},
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			gh := github.New()

			if isEnvEnabled("IP_FETCHER_MOCK_GITHUB") {
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

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        providerName,
				DefaultFileName: fileName,
				Data:            data,
			})
		},
	}
}
