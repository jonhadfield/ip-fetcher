package googleutf_test

import (
	"fmt"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/googleutf"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(googleutf.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/user-triggered-fetchers.json")

	sc := googleutf.New()
	gock.InterceptClient(sc.Client.HTTPClient)

	doc, err := sc.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, googleutf.IPv4Entry{netip.MustParsePrefix("35.187.132.96/27")})
	require.NotEmpty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv6Prefixes, googleutf.IPv6Entry{netip.MustParsePrefix("2404:f340:4010:4000::/64")})
}
