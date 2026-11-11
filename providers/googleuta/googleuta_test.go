package googleuta_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/googleuta"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(googleuta.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/user-triggered-agents.json")

	sc := googleuta.New()
	gock.InterceptClient(sc.Client.HTTPClient)

	doc, err := sc.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, googleuta.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("136.121.16.0/24")})
	require.NotEmpty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv6Prefixes, googleuta.IPv6Entry{IPv6Prefix: netip.MustParsePrefix("2001:4860:c::/124")})
}
