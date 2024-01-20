package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/azure"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func azureCmd() *cli.Command {
	const (
		// testAzureDownloadURL     = "https://download.microsoft.com/download/7/1/D/71D86715-5596-4529-9B13-DA13A5DE5B63/ServiceTags_Public_20000000.json"
		testAzureDownloadURL     = azure.WorkaroundDownloadURL
		testAzureInitialFilePath = "../../providers/azure/testdata/initial.html"
		testAzureDataFilePath    = "../../providers/azure/testdata/ServiceTags_Public_20221212.json"
		providerName             = "azure"
		fileName                 = "ServiceTags_Public.json"
	)
	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Azure prefixes",
		Usage:     "Microsoft Azure",
		UsageText: "ip-fetcher azure {--stdout | --path FILE}",
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
			path := strings.TrimSpace(c.String("path"))
			if path == "" && !c.Bool("stdout") {
				_ = cli.ShowSubcommandHelp(c)

				fmt.Println("\nerror: must specify at least one of stdout and path")

				os.Exit(1)
			}

			a := azure.New()

			if os.Getenv("IP_FETCHER_MOCK_AZURE") == "true" {
				defer gock.Off()
				u, _ := url.Parse(azure.InitialURL)
				gock.New(azure.InitialURL).
					Get(u.Path).
					Reply(200).
					File(testAzureInitialFilePath)

				uDownload, _ := url.Parse(testAzureDownloadURL)
				gock.New(testAzureDownloadURL).
					Get(uDownload.Path).
					Reply(200).
					File(testAzureDataFilePath)

				gock.InterceptClient(a.Client.HTTPClient)
			}

			data, _, _, err := a.FetchData()
			if err != nil {
				return err
			}

			if path != "" {
				var out string

				if out, err = saveFile(saveFileInput{
					provider:        providerName,
					data:            data,
					path:            path,
					defaultFileName: fileName,
				}); err != nil {
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
