package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

var version, versionOutput, tag, sha, buildDate string

const (
	autoBackup      = true
	defaultLogLevel = "info"
)

func init() {
	lvl, ok := os.LookupEnv("PF_LOG")
	if !ok {
		lvl = defaultLogLevel
	}

	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.InfoLevel
	}

	// logrus.SetFormatter(&nested.Formatter{
	// 	HideKeys:    true,
	// 	FieldsOrder: []string{"component", "category"},
	// })

	logrus.SetLevel(ll)
}
func main() {
	if tag != "" && buildDate != "" {
		versionOutput = fmt.Sprintf("[%s-%s] %s UTC", tag, sha, buildDate)
	} else {
		versionOutput = version
	}

	app := cli.NewApp()
	app.EnableBashCompletion = true

	app.Name = "prefix-fetcher"
	app.Version = versionOutput
	app.Compiled = time.Now()
	app.Authors = []*cli.Author{
		{
			Name:  "Jon Hadfield",
			Email: "jon@lessknown.co.uk",
		},
	}
	app.HelpName = ""
	app.Description = "prefix-fetcher is a tool to download and display network prefixes from various service providers."
	app.Usage = "prefix-fetcher [global options] provider [command options]"
	app.Flags = []cli.Flag{}
	app.Commands = []*cli.Command{
		abuseipdbCmd(),
		awsCmd(),
		azureCmd(),
		digitaloceanCmd(),
		gcpCmd(),
		geoipCmd(),
	}

	if err := app.Run(os.Args); err != nil {

		fmt.Printf("\nerror: %s\n", err.Error())

	}
}
