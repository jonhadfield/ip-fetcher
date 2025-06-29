package fastly_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/fastly"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(fastly.DownloadURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/fastly.json")

	cf := fastly.New()
	gock.InterceptClient(cf.Client.HTTPClient)

	ips, err := cf.Fetch()
	require.NoError(t, err)
	require.Len(t, ips.IPv4Prefixes, 19)
	require.Len(t, ips.IPv6Prefixes, 2)
	require.Contains(t, ips.IPv6Prefixes, netip.MustParsePrefix("2a04:4e40::/32"))
	require.Contains(t, ips.IPv4Prefixes, netip.MustParsePrefix("199.27.72.0/21"))
}
