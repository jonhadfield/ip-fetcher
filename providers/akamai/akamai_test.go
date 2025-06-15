package akamai_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/akamai"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/prefixes.txt")
	require.NoError(t, err)
	require.NotEmpty(t, data)

	prefixes, err := akamai.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, prefixes, 2)
	require.Contains(t, prefixes, netip.MustParsePrefix("203.0.113.0/24"))
	require.Contains(t, prefixes, netip.MustParsePrefix("2001:db8::/32"))
}

func TestFetch(t *testing.T) {
	u, err := url.Parse(akamai.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/prefixes.txt")

	a := akamai.New()
	gock.InterceptClient(a.Client.HTTPClient)

	prefixes, err := a.Fetch()
	require.NoError(t, err)
	require.Len(t, prefixes, 2)
	require.Contains(t, prefixes, netip.MustParsePrefix("203.0.113.0/24"))
	require.Contains(t, prefixes, netip.MustParsePrefix("2001:db8::/32"))
}
