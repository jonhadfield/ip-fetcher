package main

import (
	"fmt"
	"github.com/jonhadfield/ip-fetcher/providers/cloudflare"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	providerName  = "cloudflare"
	ipsv4Filename = "ips-v4"
	ipsv6Filename = "ips-v6"
)

func CloudflareCmd() *cli.Command {
	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Cloudflare ip ranges",
		UsageText: "ip-fetcher cloudflare [-4 ipv4] [-6 ipv6] {--stdout | --path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			// nolint:errcheck
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "where to save the file(s)", TakesFile: true,
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: "write to stdout", Aliases: []string{"s"},
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
			path := strings.TrimSpace(c.String("path"))

			four := c.Bool("4")
			six := c.Bool("6")
			stdOut := c.Bool("stdout")
			var msg string

			if path == "" && !stdOut {
				// nolint:errcheck
				_ = cli.ShowSubcommandHelp(c)

				fmt.Println("\nerror: must specify at least one of stdout and path")
				os.Exit(1)
			}

			cf := cloudflare.New()

			if os.Getenv("IP_FETCHER_MOCK_CLOUDFLARE") == "true" {
				u4, _ := url.Parse(cloudflare.DefaultIPv4URL)
				u6, _ := url.Parse(cloudflare.DefaultIPv6URL)

				defer gock.Off()

				url4Base := fmt.Sprintf("%s://%s", u4.Scheme, u4.Host)
				exTimeStamp := "Tue, 13 Dec 2022 06:50:50 GMT"
				gock.New(url4Base).
					Get(u4.Path).
					Reply(200).
					AddHeader("Last-Modified", exTimeStamp).
					File("../../providers/cloudflare/testdata/ips-v4")
				url6Base := fmt.Sprintf("%s://%s", u6.Scheme, u6.Host)
				gock.New(url6Base).
					Get(u6.Path).
					Reply(200).
					AddHeader("Last-Modified", exTimeStamp).
					File("../../providers/cloudflare/testdata/ips-v6")

				gock.InterceptClient(cf.Client.HTTPClient)
			}

			if four || (!four && !six) {
				dataFour, _, _, err := cf.FetchIPv4Data()
				if err != nil {
					return err
				}
				if stdOut {
					fmt.Printf("%s\n", dataFour)
				}
				if path != "" {
					p := filepath.Join(path, ipsv4Filename)

					var out string
					if out, err = saveFile(saveFileInput{
						provider: providerName,
						data:     dataFour,
						path:     p,
					}); err != nil {
						return err
					}

					msg = fmt.Sprintf("ipv4 data written to %s\n", out)
				}
			}

			if six || (!four && !six) {
				dataSix, _, _, err := cf.FetchIPv6Data()
				if err != nil {
					return err
				}

				if stdOut {
					fmt.Printf("%s\n", dataSix)
				}

				if path != "" {
					p := filepath.Join(path, ipsv6Filename)

					var out string
					if out, err = saveFile(saveFileInput{
						provider: providerName,
						data:     dataSix,
						path:     p,
					}); err != nil {
						return err
					}

					msg += fmt.Sprintf("ipv6 data written to %s\n", out)
				}
			}

			if msg != "" {
				if stdOut {
					fmt.Println()
				}

				_, _ = os.Stderr.WriteString(msg)
			}

			return nil
		},
	}
}
