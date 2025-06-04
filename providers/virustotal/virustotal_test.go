package virustotal_test

import (
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/virustotal"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/ip-addresses.txt")
	require.NoError(t, err)

	prefixes, err := virustotal.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, prefixes, 2)
	require.Contains(t, prefixes, netip.MustParsePrefix("198.51.100.0/24"))
	require.Contains(t, prefixes, netip.MustParsePrefix("2001:db8:abcd::/48"))
}

func TestFetch(t *testing.T) {
	u, err := url.Parse(virustotal.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/ip-addresses.txt")

	v := virustotal.New()
	gock.InterceptClient(v.Client.HTTPClient)

	prefixes, err := v.Fetch()
	require.NoError(t, err)
	require.Len(t, prefixes, 2)
	require.Contains(t, prefixes, netip.MustParsePrefix("198.51.100.0/24"))
	require.Contains(t, prefixes, netip.MustParsePrefix("2001:db8:abcd::/48"))
}
