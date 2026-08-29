package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/zoom"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func zoomCmd() *cli.Command {
	const (
		providerName  = "zoom"
		fileNameData  = "zoom.txt"
		fileNameLines = "zoom-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Zoom service ranges",
		Usage:     "Zoom (meeting and phone service ranges)",
		UsageText: "ip-fetcher zoom {--stdout | --Path FILE} [--lines]",
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

			p := zoom.New()

			if isEnvEnabled("IP_FETCHER_MOCK_ZOOM") {
				defer gock.Off()

				u, _ := url.Parse(zoom.DownloadURL)
				gock.New(zoom.DownloadURL).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/zoom/testdata/zoom.txt")
				gock.InterceptClient(p.Client.HTTPClient)
			}

			data, err := zoomData(c, &p)
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

// zoomData returns newline separated prefixes when --lines is set, and the
// upstream document otherwise.
func zoomData(c *cli.Context, p *zoom.Zoom) ([]byte, error) {
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
