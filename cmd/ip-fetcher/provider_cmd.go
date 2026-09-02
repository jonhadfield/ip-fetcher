package main

import (
	"net/http"

	"github.com/urfave/cli/v2"
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
