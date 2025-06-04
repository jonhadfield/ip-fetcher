package zscaler_test

import (
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/zscaler"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/prefixes.txt")
	require.NoError(t, err)
	require.NotEmpty(t, data)

	prefixes, err := zscaler.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, prefixes, 2)
	require.Contains(t, prefixes, netip.MustParsePrefix("198.51.100.0/24"))
	require.Contains(t, prefixes, netip.MustParsePrefix("2001:db8:2::/48"))
}

func TestFetch(t *testing.T) {
	u, err := url.Parse(zscaler.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/prefixes.txt")

	z := zscaler.New()
	gock.InterceptClient(z.Client.HTTPClient)

	prefixes, err := z.Fetch()
	require.NoError(t, err)
	require.Len(t, prefixes, 2)
	require.Contains(t, prefixes, netip.MustParsePrefix("198.51.100.0/24"))
	require.Contains(t, prefixes, netip.MustParsePrefix("2001:db8:2::/48"))
}
