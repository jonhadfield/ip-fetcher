package google

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
		File("testdata/goog.json")

	ac := New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, IPv4Entry{netip.MustParsePrefix("34.2.0.0/16")})
	require.NotEmpty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv6Prefixes, IPv6Entry{netip.MustParsePrefix("2620:120:e001::/40")})
}
