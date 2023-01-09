package main

import (
	"github.com/jonhadfield/prefix-fetcher/aws"
	"github.com/urfave/cli/v2"
)

func awsCmd() *cli.Command {
	return &cli.Command{
		Name:  "aws",
		Usage: "fetch aws prefixes",
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
			a := aws.New()
			data, _, _, err := a.FetchData()
			if err != nil {
				return err
			}

			path := c.String("path")
			if path != "" {
				return saveFile(saveFileInput{
					provider:        "aws",
					data:            data,
					path:            path,
					defaultFileName: "ip-ranges.json",
				})
			}

			return nil
		},
	}
}
