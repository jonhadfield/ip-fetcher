package googlebot_test

import (
	"fmt"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/googlebot"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(googlebot.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/googlebot.json")

	ac := googlebot.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, googlebot.IPv4Entry{netip.MustParsePrefix("66.249.79.64/27")})
	require.NotEmpty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv6Prefixes, googlebot.IPv6Entry{netip.MustParsePrefix("2001:4860:4801:11::/64")})
}
