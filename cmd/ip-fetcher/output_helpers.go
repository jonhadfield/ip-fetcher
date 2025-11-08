package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

func resolveOutputTargets(c *cli.Context) (string, bool, error) {
	path := strings.TrimSpace(c.String("Path"))
	stdout := c.Bool("stdout")

	if path == "" && !stdout {
		_ = cli.ShowSubcommandHelp(c)
		fmt.Println("\n" + errStdoutOrPathRequired)

		return "", false, cli.Exit(errStdoutOrPathRequired, 1)
	}

	return path, stdout, nil
}

func writeOutputs(path string, stdout bool, input SaveFileInput) error {
	if path != "" {
		input.Path = path
		out, err := SaveFile(input)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(os.Stderr, fmtDataWrittenTo, out)
	}

	if stdout {
		fmt.Printf("%s\n", input.Data)
	}

	return nil
}
