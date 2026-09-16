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
		ahrefsCmd(),
		akamaiCmd(),
		alibabaCmd(),
		amazonbotCmd(),
		anthropicCmd(),
		applebotCmd(),
		atlassianCmd(),
		awsCmd(),
		azureCmd(),
		bingbotCmd(),
		betterstackCmd(),
		blocklistdeCmd(),
		bunnyCmd(),
		cacheflyCmd(),
		cdn77Cmd(),
		ccbotCmd(),
		cinsscoreCmd(),
		checklyCmd(),
		cloudflareCmd(),
		cymruCmd(),
		contaboCmd(),
		datadogCmd(),
		detectifyCmd(),
		digitaloceanCmd(),
		dshieldCmd(),
		duckduckbotCmd(),
		emergingthreatsCmd(),
		fastlyCmd(),
		feodoCmd(),
		flyioCmd(),
		gcpCmd(),
		gcoreCmd(),
		geoipCmd(),
		githubCmd(),
		gitlabCmd(),
		googleCmd(),
		googlebotCmd(),
		googlescCmd(),
		googleutfCmd(),
		grafanaCmd(),
		greensnowCmd(),
		hetznerCmd(),
		huaweiCmd(),
		iCloudPRCmd(),
		ibmcloudCmd(),
		impervaCmd(),
		intercomCmd(),
		leasewebCmd(),
		linodeCmd(),
		m247Cmd(),
		m365Cmd(),
		mullvadCmd(),
		ociCmd(),
		newrelicCmd(),
		oktaCmd(),
		openaiCmd(),
		ovhCmd(),
		perplexitybotCmd(),
		pingdomCmd(),
		publishCmd(),
		renderCmd(),
		salesforceCmd(),
		scalewayCmd(),
		sentryCmd(),
		site24x7Cmd(),
		spamhausCmd(),
		stripeCmd(),
		statuscakeCmd(),
		tenableCmd(),
		tencentCmd(),
		torCmd(),
		updownCmd(),
		uptimerobotCmd(),
		uptrendsCmd(),
		urlCmd(),
		zoomCmd(),
		vultrCmd(),
		zscalerCmd(),
	}

	return app
}
