package main_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	mainpkg "github.com/jonhadfield/ip-fetcher/cmd/ip-fetcher"
	"github.com/stretchr/testify/require"
)

func testCleanUp(args []string) {
	os.Args = args
	_ = os.Unsetenv("TEST_EXIT")
}

// captureStdout runs the CLI with the given arguments and returns whatever it
// wrote to stdout.
func captureStdout(t *testing.T, args []string) string {
	t.Helper()

	old := os.Stdout

	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	app := mainpkg.GetApp()
	os.Args = args
	runErr := app.Run(os.Args)

	_ = w.Close()
	os.Stdout = old

	require.NoError(t, runErr)

	return <-outC
}
