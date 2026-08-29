package main

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/openai"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func openaiCmd() *cli.Command {
	const (
		providerName  = "openai"
		fileName      = "openai.json"
		fileNameLines = "openai-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch OpenAI bot prefixes",
		Usage:     "OpenAI Bots (GPTBot, OAI-SearchBot and ChatGPT-User)",
		UsageText: "ip-fetcher openai {--stdout | --Path FILE} [--lines]",
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

			a := openai.New()

			if isEnvEnabled("IP_FETCHER_MOCK_OPENAI") {
				defer gock.Off()

				for downloadURL, testDataFile := range map[string]string{
					openai.GPTBotDownloadURL:      "../../providers/openai/testdata/gptbot.json",
					openai.SearchBotDownloadURL:   "../../providers/openai/testdata/searchbot.json",
					openai.ChatGPTUserDownloadURL: "../../providers/openai/testdata/chatgpt-user.json",
				} {
					u, _ := url.Parse(downloadURL)
					gock.New(downloadURL).
						Get(u.Path).
						Reply(http.StatusOK).
						File(testDataFile)
				}

				gock.InterceptClient(a.Client.HTTPClient)
			}

			doc, err := a.Fetch()
			if err != nil {
				return err
			}

			var data []byte
			if c.Bool(formatLines) {
				if data, err = docToLines(doc); err != nil {
					return err
				}
			} else {
				data, err = json.MarshalIndent(doc, "", "  ")
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
