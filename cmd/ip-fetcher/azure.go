package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

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
				Name:  "lines",
				Usage: usageLinesOutput,
			},
		},
		Action: func(c *cli.Context) error {
			path := strings.TrimSpace(c.String("Path"))
			if path == "" && !c.Bool("stdout") {
				_ = cli.ShowSubcommandHelp(c)

				fmt.Println("\n" + errStdoutOrPathRequired) //nolint:forbidigo

				os.Exit(1)
			}

			a := azure.New()

			if os.Getenv("IP_FETCHER_MOCK_AZURE") == "true" {
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
			var err error
			if c.Bool("lines") {
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

			if path != "" {
				var out string

				df := fileName
				if c.Bool("lines") {
					df = fileNameLines
				}
				if out, err = SaveFile(SaveFileInput{
					Provider:        providerName,
					Data:            data,
					Path:            path,
					DefaultFileName: df,
				}); err != nil {
					return err
				}

				_, _ = fmt.Fprintf(os.Stderr, fmtDataWrittenTo, out)
			}

			if c.Bool("stdout") {
				fmt.Printf("%s\n", data)
			}

			return nil
		},
	}
}
