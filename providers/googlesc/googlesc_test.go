package googlesc_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/googlesc"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(googlesc.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/special-crawlers.json")

	sc := googlesc.New()
	gock.InterceptClient(sc.Client.HTTPClient)

	doc, err := sc.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, googlesc.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("209.85.238.128/27")})
	require.NotEmpty(t, doc.IPv6Prefixes)
	require.Contains(
		t,
		doc.IPv6Prefixes,
		googlesc.IPv6Entry{IPv6Prefix: netip.MustParsePrefix("2001:4860:4801:2092::/64")},
	)
}
