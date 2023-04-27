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
	app.HelpName = ""
	app.Description = "ip-fetcher is a tool to download and display network prefixes from various service providers."
	app.Usage = "ip-fetcher [global options] provider [command options]"
	app.Flags = []cli.Flag{}
	app.Commands = []*cli.Command{
		abuseipdbCmd(),
		awsCmd(),
		azureCmd(),
		digitaloceanCmd(),
		gcpCmd(),
		geoipCmd(),
	}

	// app.Run(os.Args)

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\nerror: %s\n", err.Error())
	}
}
