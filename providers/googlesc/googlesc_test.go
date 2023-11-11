package googlesc

import (
	"fmt"
	"net/netip"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/special-crawlers.json")

	sc := New()
	gock.InterceptClient(sc.Client.HTTPClient)

	doc, err := sc.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, IPv4Entry{netip.MustParsePrefix("209.85.238.128/27")})
	require.NotEmpty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv6Prefixes, IPv6Entry{netip.MustParsePrefix("2001:4860:4801:2092::/64")})
}
