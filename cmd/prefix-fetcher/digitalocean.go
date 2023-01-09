package main

import (
	"github.com/jonhadfield/prefix-fetcher/digitalocean"
	"github.com/urfave/cli/v2"
)

func digitaloceanCmd() *cli.Command {
	return &cli.Command{
		Name:  "digitalocean",
		Usage: "fetch digitalocean prefixes",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "where to save the file", Aliases: []string{"p"},
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "output as yaml or json", Aliases: []string{"f"},
			},
		},
		Action: func(c *cli.Context) error {
			a := digitalocean.New()
			data, _, _, err := a.FetchData()
			if err != nil {
				return err
			}

			path := c.String("path")
			if path != "" {
				return saveFile(saveFileInput{
					provider:        "digitalocean",
					data:            data,
					path:            path,
					defaultFileName: "ServiceTags_Public.json",
				})
			}

			return nil
		},
	}
}
