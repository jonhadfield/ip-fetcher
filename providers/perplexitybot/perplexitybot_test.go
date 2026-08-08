package perplexitybot_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/perplexitybot"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(perplexitybot.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/perplexitybot.json")

	ac := perplexitybot.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, perplexitybot.IPv4Entry{netip.MustParsePrefix("107.20.236.150/32")})
	require.Contains(t, doc.IPv4Prefixes, perplexitybot.IPv4Entry{netip.MustParsePrefix("18.97.9.96/29")})
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/perplexitybot.json")
	require.NoError(t, err)

	doc, err := perplexitybot.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 4)
	require.Empty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv4Prefixes, perplexitybot.IPv4Entry{netip.MustParsePrefix("3.224.62.45/32")})
	require.Equal(t, 2025, doc.CreationTime.Year())
}
