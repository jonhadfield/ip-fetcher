package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/aws"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

const (
	awsProviderName  = "aws"
	awsFileName      = "ip-ranges.json"
	awsFileNameLines = "aws-prefixes.txt"
)

func awsCmd() *cli.Command {
	return &cli.Command{
		Name:      awsProviderName,
		HelpName:  "- fetch AWS prefixes",
		Usage:     "Amazon Web Services",
		UsageText: "ip-fetcher aws {--stdout | --Path FILE} [--lines]",
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

			a := aws.New()

			mockEnabled, err := configureAWSMock(&a)
			if err != nil {
				return err
			}
			if mockEnabled {
				defer gock.Off()
			}

			data, fileName, err := awsData(&a, c.Bool(formatLines))
			if err != nil {
				return err
			}

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        awsProviderName,
				DefaultFileName: fileName,
				Data:            data,
			})
		},
	}
}

func configureAWSMock(a *aws.AWS) (bool, error) {
	if !isEnvEnabled("IP_FETCHER_MOCK_AWS") {
		return false, nil
	}

	urlBase := aws.DownloadURL
	u, err := url.Parse(urlBase)
	if err != nil {
		return false, err
	}

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("../../providers/aws/testdata/ip-ranges.json")

	gock.InterceptClient(a.Client.HTTPClient)

	return true, nil
}

func awsData(a *aws.AWS, asLines bool) ([]byte, string, error) {
	if asLines {
		doc, _, err := a.Fetch()
		if err != nil {
			return nil, "", err
		}

		data, err := docToLines(doc)
		if err != nil {
			return nil, "", err
		}

		return data, awsFileNameLines, nil
	}

	data, _, _, err := a.FetchData()
	if err != nil {
		return nil, "", err
	}

	return data, awsFileName, nil
}
