package commands

import (
	"fmt"
	"github.com/jonhadfield/prefix-fetcher/cloudflare"
	"github.com/urfave/cli/v2"
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
		UsageText: "prefix-fetcher cloudflare [ -4 ipv4] [ -6 ipv6] {--stdout | --path FILE}",
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
					if err = saveFile(saveFileInput{
						provider: providerName,
						data:     dataFour,
						path:     p,
					}); err != nil {
						return err
					}

					var ap string
					ap, err = filepath.Abs(p)
					if err != nil {
						return err
					}
					msg = fmt.Sprintf("ipv4 data written to %s\n", ap)
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
					if err = saveFile(saveFileInput{
						provider: providerName,
						data:     dataSix,
						path:     p,
					}); err != nil {
						return err
					}

					var ap string
					ap, err = filepath.Abs(p)
					if err != nil {
						return err
					}

					msg += fmt.Sprintf("ipv6 data written to %s\n", ap)
				}
			}

			if msg != "" {
				fmt.Printf("\n%s", msg)
			}

			return nil
		},
	}
}
