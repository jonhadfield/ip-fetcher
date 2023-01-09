package main

import (
	"github.com/jonhadfield/prefix-fetcher/gcp"
	"github.com/urfave/cli/v2"
)

func gcpCmd() *cli.Command {
	return &cli.Command{
		Name:  "gcp",
		Usage: "fetch gcp prefixes",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "where to save the file", Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  "format",
				Usage: "output as yaml or json", Aliases: []string{"f"},
			},
		},
		Action: func(c *cli.Context) error {
			a := gcp.New()
			data, _, _, err := a.FetchData()
			if err != nil {
				return err
			}

			path := c.String("path")
			if path != "" {
				return saveFile(saveFileInput{
					provider:        "gcp",
					data:            data,
					path:            path,
					defaultFileName: "cloud.json",
				})
			}

			return nil
		},
	}
}
