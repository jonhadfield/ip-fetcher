package m247_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/m247"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(fmt.Sprintf(m247.DownloadURL, m247.ASNs[0]))
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/prefixes.json")

	o := m247.New()
	gock.InterceptClient(o.Client.HTTPClient)

	doc, err := o.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("45.141.155.0/24"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2a0a:c201::/33"))
}
