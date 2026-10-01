package main_test

import (
	"os"
	"testing"

	mainpkg "github.com/jonhadfield/ip-fetcher/cmd/ip-fetcher"
	"github.com/stretchr/testify/require"
)

func TestAggregateRequiresLines(t *testing.T) {
	defer testCleanUp(os.Args)

	t.Setenv("IP_FETCHER_MOCK_AWS", "true")

	app := mainpkg.GetApp()
	os.Args = []string{"ip-fetcher", "aws", "--stdout", "--aggregate", "exact"}
	err := app.Run(os.Args)
	require.Error(t, err)
	require.Contains(t, err.Error(), "--aggregate requires --lines")
}
