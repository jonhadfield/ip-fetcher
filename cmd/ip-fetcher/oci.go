package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/oci"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

const sOCI = "oci"

func ociCmd() *cli.Command {
	const (
		providerName = sOCI
		fileName     = "public_ip_ranges.json"
	)

	return &cli.Command{
		Name:      providerName,
		Usage:     "Oracle Cloud Infrastructure",
		HelpName:  "- fetch OCI (Oracle Cloud Infrastructure) prefixes",
		UsageText: "ip-fetcher oci {--stdout | --Path FILE}",
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
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			a := oci.New()

			if isEnvEnabled("IP_FETCHER_MOCK_OCI") {
				defer gock.Off()
				urlBase := oci.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/oci/testdata/public_ip_ranges.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			data, _, _, err := a.FetchData()
			if err != nil {
				return err
			}

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        providerName,
				DefaultFileName: fileName,
				Data:            data,
			})
		},
	}
}
