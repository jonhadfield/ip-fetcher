package main

import (
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/maxmind/geoip"
	"github.com/urfave/cli/v2"
)

func geoipCmd() *cli.Command {
	return &cli.Command{
		Name:      "geoip",
		HelpName:  "- fetch MaxMind GeoIP prefixes",
		Usage:     "MaxMind GeoIP",
		UsageText: "ip-fetcher geoip --key=mykey --Path=mypath [ --format=(csv | mmdb) ] [ --edition=(GeoLite2 | GeoIP) ] [ --extract ]",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "key",
				Usage: "license key", Aliases: []string{"k"}, Required: true,
			},
			&cli.StringFlag{
				Name:  "Path",
				Usage: "where to save the files", Aliases: []string{"p"}, Required: true,
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "database format csv or mmdb", Value: "csv", Aliases: []string{"f"},
			},
			&cli.StringFlag{
				Name:  "edition",
				Usage: "GeoLite2 or GeoIP2", Value: "GeoLite2", Aliases: []string{"e"},
			},
			&cli.BoolFlag{
				Name:  "extract",
				Usage: "extract compressed downloads", Value: true,
			},
		},
		Action: func(c *cli.Context) error {
			a := geoip.New()
			a.LicenseKey = c.String("key")
			a.Edition = c.String("edition")
			a.DBFormat = c.String("format")
			a.Root = strings.TrimSpace(c.String("Path"))
			a.Extract = c.Bool("extract")
			_, err := a.FetchFiles(geoip.FetchFilesInput{
				ASN:     true,
				Country: true,
				City:    true,
			})
			if err != nil {
				return err
			}

			if a.Extract {
				return nil
			}

			return nil
		},
	}
}
