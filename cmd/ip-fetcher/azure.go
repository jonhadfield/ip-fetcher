package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/azure"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func azureCmd() *cli.Command {
	const (
		testMockAzureDownloadURL = azure.WorkaroundDownloadURL
		// testMockAzureDownloadURL = "https://download.microsoft.com/download/7/1/D/71D86715-5596-4529-9B13-DA13A5DE5B63/ServiceTags_Public_2000000.json"
		// testMockAzureDownloadURL     = azure.WorkaroundDownloadURL
		testAzureInitialFilePath = "../../providers/azure/testdata/initial.html"
		testAzureDataFilePath    = "../../providers/azure/testdata/ServiceTags_Public_20221212.json"
		providerName             = "azure"
		fileName                 = "ServiceTags_Public.json"
		fileNameLines            = "azure-prefixes.txt"
	)
	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Azure prefixes",
		Usage:     "Microsoft Azure",
		UsageText: "ip-fetcher azure {--stdout | --Path FILE} [--lines]",
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

			a := azure.New()

			if isEnvEnabled("IP_FETCHER_MOCK_AZURE") {
				defer gock.Off()
				// u, _ := url.Parse(azure.InitialURL)
				// gock.New(azure.InitialURL).
				// 	Get(u.Path).
				// 	Reply(http.StatusOK).
				// 	File(testAzureInitialFilePath)

				uDownload, _ := url.Parse(testMockAzureDownloadURL)
				gock.New(testMockAzureDownloadURL).
					Get(uDownload.Path).
					Reply(http.StatusOK).
					File(testAzureDataFilePath)

				gock.InterceptClient(a.Client.HTTPClient)
			}

			var data []byte
			if c.Bool(formatLines) {
				var doc azure.Doc
				if doc, _, err = a.Fetch(); err != nil {
					return err
				}
				if data, err = docToLines(doc); err != nil {
					return err
				}
			} else {
				data, _, _, err = a.FetchData()
				if err != nil {
					return err
				}
			}

			defaultName := fileName
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
