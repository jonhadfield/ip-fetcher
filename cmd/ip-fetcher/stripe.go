package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/stripe"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/yaml.v3"
)

const (
	providerNameStripe   = "stripe"
	fileNameOutputStripe = "stripe.json"
	fileNameLinesStripe  = "stripe-prefixes.txt"
)

var stripeFormats = []string{formatJSON, formatYAML, formatLines}

func stripeCmd() *cli.Command {
	return &cli.Command{
		Name:      providerNameStripe,
		HelpName:  "- fetch Stripe prefixes",
		Usage:     "Stripe",
		UsageText: "ip-fetcher stripe {--stdout | --Path FILE} [--lines]",
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
			&cli.StringFlag{
				Name:  flagFormat,
				Usage: strings.Join(stripeFormats, ", "), Value: formatJSON, Aliases: []string{"f"},
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

			a := stripe.New()

			if isEnvEnabled("IP_FETCHER_MOCK_STRIPE") {
				defer gock.Off()

				wh, _ := url.Parse(stripe.WebhooksURL)
				gock.New(fmt.Sprintf("%s://%s", wh.Scheme, wh.Host)).
					Get(wh.Path).
					Reply(http.StatusOK).
					File("../../providers/stripe/testdata/ips_webhooks.json")

				api, _ := url.Parse(stripe.APIURL)
				gock.New(fmt.Sprintf("%s://%s", api.Scheme, api.Host)).
					Get(api.Path).
					Reply(http.StatusOK).
					File("../../providers/stripe/testdata/ips_api.json")

				gock.InterceptClient(a.Client.HTTPClient)
			}

			var doc stripe.Doc
			if doc, err = a.Fetch(); err != nil {
				return err
			}

			format := c.String(flagFormat)
			if c.Bool(formatLines) {
				format = formatLines
			}

			return stripeOutput(doc, format, stdout, path)
		},
	}
}

func stripeOutput(doc stripe.Doc, format string, stdout bool, path string) error {
	if !slices.Contains(stripeFormats, format) {
		return fmt.Errorf("invalid format: %s\n       choose from: %s",
			format, strings.Join(stripeFormats, ", "))
	}

	var (
		data []byte
		err  error
	)

	switch format {
	case formatLines:
		data = prefixesToLines(doc.IPv4Prefixes, doc.IPv6Prefixes)
	case formatYAML:
		if data, err = yaml.Marshal(doc); err != nil {
			return err
		}
	case formatJSON:
		if data, err = json.MarshalIndent(doc, "", " "); err != nil {
			return err
		}
	}

	if stdout {
		fmt.Printf("%s\n\n", data)
	}

	defaultName := fileNameOutputStripe
	if format == formatLines {
		defaultName = fileNameLinesStripe
	}

	return writeOutputs(path, false, SaveFileInput{
		Provider:        providerNameStripe,
		DefaultFileName: defaultName,
		Data:            data,
	})
}
