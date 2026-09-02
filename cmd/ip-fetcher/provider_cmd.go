package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

// The helpers here hold the parts every provider command repeats: the same
// three flags, the same usage error handling, the same choice between the
// upstream document and newline separated prefixes, and the same output naming.

func onUsageError(cCtx *cli.Context, err error, _ bool) error {
	_ = cli.ShowSubcommandHelp(cCtx)

	return err
}

func providerFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  flagPath,
			Usage: usageWhereToSaveFile, Aliases: []string{"p"},
		},
		&cli.BoolFlag{
			Name:  flagStdout,
			Usage: usageWriteToStdout, Aliases: []string{"s"},
		},
		&cli.BoolFlag{
			Name:  formatLines,
			Usage: usageLinesOutput,
		},
	}
}

// mockSource serves one of a provider's URLs from a testdata file, so the CLI
// tests can exercise a command without reaching the network.
type mockSource struct {
	url  string
	file string
}

// providerSpec describes a provider command: what it is called, what it writes,
// and how to fetch it.
type providerSpec struct {
	name      string
	helpName  string
	usage     string
	dataFile  string
	linesFile string
	// mockEnv names the variable that serves the provider from its testdata.
	mockEnv string
	mocks   []mockSource
	// newProvider returns the two fetches and the client the mocks intercept.
	newProvider func() (
		fetchData func() ([]byte, http.Header, int, error),
		fetchDoc func() (any, error),
		client *http.Client,
	)
}

// providerCommand builds the command for a provider that needs nothing beyond
// fetching its document and writing it out.
func providerCommand(spec providerSpec) *cli.Command {
	return &cli.Command{
		Name:         spec.name,
		HelpName:     "- fetch " + spec.helpName,
		Usage:        spec.usage,
		UsageText:    fmt.Sprintf("ip-fetcher %s {--stdout | --Path FILE} [--lines]", spec.name),
		OnUsageError: onUsageError,
		Flags:        providerFlags(),
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			fetchData, fetchDoc, client := spec.newProvider()

			if isEnvEnabled(spec.mockEnv) {
				defer gock.Off()

				for _, m := range spec.mocks {
					u, _ := url.Parse(m.url)
					gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
						Get(u.Path).
						Reply(http.StatusOK).
						File(m.file)
				}

				gock.InterceptClient(client)
			}

			data, err := providerData(c, fetchData, fetchDoc)
			if err != nil {
				return err
			}

			return writeProviderOutputs(c, path, stdout, spec.name, spec.dataFile, spec.linesFile, data)
		},
	}
}

// providerData returns newline separated prefixes when --lines is set, and the
// upstream document otherwise.
func providerData(
	c *cli.Context,
	fetchData func() ([]byte, http.Header, int, error),
	fetchDoc func() (any, error),
) ([]byte, error) {
	if c.Bool(formatLines) {
		doc, err := fetchDoc()
		if err != nil {
			return nil, err
		}

		return docToLines(doc)
	}

	data, _, _, err := fetchData()

	return data, err
}

// writeProviderOutputs saves the document under the name the chosen format
// implies, unless the user named a file themselves.
func writeProviderOutputs(c *cli.Context, path string, stdout bool, provider, dataFile, linesFile string, data []byte) error {
	defaultName := dataFile
	if c.Bool(formatLines) {
		defaultName = linesFile
	}

	return writeOutputs(path, stdout, SaveFileInput{
		Provider:        provider,
		DefaultFileName: defaultName,
		Data:            data,
	})
}
