package main

import (
	"fmt"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

var version, versionOutput, tag, sha, buildDate string

func main() {
	pflog.SetLogLevel()

	if tag != "" && buildDate != "" {
		versionOutput = fmt.Sprintf("[%s-%s] %s UTC", tag, sha, buildDate)
	} else {
		versionOutput = version
	}

	app := getApp()

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\nerror: %s\n", err.Error())
	}
}

func getApp() *cli.App {
	app := cli.NewApp()

	app.EnableBashCompletion = true
	app.Name = "ip-fetcher"
	app.Version = versionOutput
	app.Compiled = time.Now()
	app.Authors = []*cli.Author{
		{
			Name:  "Jon Hadfield",
			Email: "jon@lessknown.co.uk",
		},
	}
	app.Usage = "Download and display ips for various cloud providers and services"
	app.Commands = []*cli.Command{
		abuseipdbCmd(),
		awsCmd(),
		azureCmd(),
		cloudflareCmd(),
		digitaloceanCmd(),
		gcpCmd(),
		geoipCmd(),
		googleCmd(),
		ociCmd(),
		urlCmd(),
	}

	return app
}
