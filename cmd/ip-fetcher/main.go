package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/urfave/cli/v2"
)

var version, versionOutput, tag, sha, buildDate string

func main() {
	pflog.SetLogLevel()

	if tag != "" && buildDate != "" {
		versionOutput = fmt.Sprintf("[%s-%s] %s UTC", tag, sha, buildDate)
	} else {
		versionOutput = version
	}

	app := GetApp()

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\nerror: %s\n", err.Error())
	}
}

func GetApp() *cli.App {
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
		akamaiCmd(),
		alibabaCmd(),
		awsCmd(),
		azureCmd(),
		bingbotCmd(),
		cloudflareCmd(),
		digitaloceanCmd(),
		fastlyCmd(),
		gcpCmd(),
		geoipCmd(),
		githubCmd(),
		googleCmd(),
		googlebotCmd(),
		googlescCmd(),
		googleutfCmd(),
		hetznerCmd(),
		iCloudPRCmd(),
		linodeCmd(),
		m247Cmd(),
		ociCmd(),
		ovhCmd(),
		publishCmd(),
		scalewayCmd(),
		urlCmd(),
		vultrCmd(),
		zscalerCmd(),
	}

	return app
}
