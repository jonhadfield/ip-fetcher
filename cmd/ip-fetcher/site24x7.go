package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/site24x7"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func site24x7Cmd() *cli.Command {
	const (
		providerName  = "site24x7"
		fileNameData  = "site24x7.json"
		fileNameLines = "site24x7-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Site24x7 monitoring location addresses",
		Usage:     "Site24x7 (monitoring locations)",
		UsageText: "ip-fetcher site24x7 {--stdout | --Path FILE} [--lines]",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagPath,
				Usage: usageWhereToSaveFile, Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  flagStdout,
				Usage: usageWriteToStdout, Aliases: []string{"s"},
			},
			&cli.BoolFlag{
				Name:  formatLines,
				Usage: usageLinesOutput,
			},
		},
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			p := site24x7.New()

			if isEnvEnabled("IP_FETCHER_MOCK_SITE24X7") {
				defer gock.Off()

				u, _ := url.Parse(site24x7.LocationsURL)
				gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/site24x7/testdata/locations.html")

				// the export link the fixture page carries.
				gock.New("https://creatorapp.zohopublic.in").
					Get("/mesite24x7/location-manager/json/IP_Address_View/TESTTOKEN").
					Reply(http.StatusOK).
					File("../../providers/site24x7/testdata/locations.json")

				gock.InterceptClient(p.Client.HTTPClient)
			}

			data, err := site24x7Data(c, &p)
			if err != nil {
				return err
			}

			defaultName := fileNameData
			if c.Bool(formatLines) {
				defaultName = fileNameLines
			}

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        providerName,
				DefaultFileName: defaultName,
				Data:            data,
			})
		},
	}
}

// site24x7Data returns newline separated prefixes when --lines is set, and the
// upstream document otherwise.
func site24x7Data(c *cli.Context, p *site24x7.Site24x7) ([]byte, error) {
	if c.Bool(formatLines) {
		doc, err := p.Fetch()
		if err != nil {
			return nil, err
		}

		return docToLines(doc)
	}

	data, _, _, err := p.FetchData()

	return data, err
}
