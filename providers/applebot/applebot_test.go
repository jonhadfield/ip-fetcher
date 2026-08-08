package applebot_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/applebot"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(applebot.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/applebot.json")

	ac := applebot.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, applebot.IPv4Entry{netip.MustParsePrefix("17.241.208.160/27")})
	require.Contains(t, doc.IPv4Prefixes, applebot.IPv4Entry{netip.MustParsePrefix("17.22.237.0/24")})
	require.Empty(t, doc.IPv6Prefixes)
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/applebot.json")
	require.NoError(t, err)

	doc, err := applebot.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 4)
	require.Contains(t, doc.IPv4Prefixes, applebot.IPv4Entry{netip.MustParsePrefix("17.246.15.0/24")})
	require.Empty(t, doc.IPv6Prefixes)
	require.Equal(t, 2023, doc.CreationTime.Year())
}
