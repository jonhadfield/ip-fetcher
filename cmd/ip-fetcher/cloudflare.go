package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/cloudflare"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

const (
	providerName    = "cloudflare"
	IPsV4Filename   = "ips-v4"
	IPsV6Filename   = "ips-v6"
	messageCapacity = 2
)

func cloudflareCmd() *cli.Command { //nolint:gocognit,funlen,nestif
	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Cloudflare ip ranges",
		Usage:     "Cloudflare",
		UsageText: "ip-fetcher cloudflare [-4 ipv4] [-6 ipv6] {--stdout | --Path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			// nolint:errcheck
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "Path",
				Usage: "where to save the file(s)", TakesFile: true,
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: usageWriteToStdout, Aliases: []string{"s"},
			},
			&cli.BoolFlag{
				Name:  "4",
				Usage: "ipv4", Aliases: []string{"ipv4"},
			},
			&cli.BoolFlag{
				Name:  "6",
				Usage: "ipv4", Aliases: []string{"ipv6"},
			},
		},
		Action: func(c *cli.Context) error {
			path, stdOut, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			cf := cloudflare.New()

			if isEnvEnabled("IP_FETCHER_MOCK_CLOUDFLARE") {
				u4, _ := url.Parse(cloudflare.DefaultIPv4URL)
				u6, _ := url.Parse(cloudflare.DefaultIPv6URL)

				defer gock.Off()

				url4Base := fmt.Sprintf("%s://%s", u4.Scheme, u4.Host)
				exTimeStamp := "Tue, 13 Dec 2022 06:50:50 GMT"
				gock.New(url4Base).
					Get(u4.Path).
					Reply(http.StatusOK).
					AddHeader(web.LastModifiedHeader, exTimeStamp).
					File("../../providers/cloudflare/testdata/ips-v4")
				url6Base := fmt.Sprintf("%s://%s", u6.Scheme, u6.Host)
				gock.New(url6Base).
					Get(u6.Path).
					Reply(http.StatusOK).
					AddHeader(web.LastModifiedHeader, exTimeStamp).
					File("../../providers/cloudflare/testdata/ips-v6")

				gock.InterceptClient(cf.Client.HTTPClient)
			}

			processIPv4 := c.Bool("4")
			processIPv6 := c.Bool("6")
			if !processIPv4 && !processIPv6 {
				processIPv4 = true
				processIPv6 = true
			}

			messages := make([]string, 0, messageCapacity)

			if processIPv4 { //nolint:nestif
				ipv4Data, _, _, fetchErr := cf.FetchIPv4Data()
				if fetchErr != nil {
					return fetchErr
				}

				if stdOut {
					fmt.Printf("%s\n", ipv4Data)
				}

				if path != "" {
					savedPath, saveErr := SaveFile(SaveFileInput{
						Provider: providerName,
						Data:     ipv4Data,
						Path:     filepath.Join(path, IPsV4Filename),
					})
					if saveErr != nil {
						return saveErr
					}

					messages = append(messages, fmt.Sprintf("ipv4 Data written to %s", savedPath))
				}
			}

			if processIPv6 { //nolint:nestif
				ipv6Data, _, _, fetchErr := cf.FetchIPv6Data()
				if fetchErr != nil {
					return fetchErr
				}

				if stdOut {
					fmt.Printf("%s\n", ipv6Data)
				}

				if path != "" {
					savedPath, saveErr := SaveFile(SaveFileInput{
						Provider: providerName,
						Data:     ipv6Data,
						Path:     filepath.Join(path, IPsV6Filename),
					})
					if saveErr != nil {
						return saveErr
					}

					messages = append(messages, fmt.Sprintf("ipv6 Data written to %s", savedPath))
				}
			}

			if len(messages) > 0 {
				notification := strings.Join(messages, "\n") + "\n"
				if stdOut {
					fmt.Println()
				}
				_, _ = os.Stderr.WriteString(notification)
			}

			return nil
		},
	}
}
