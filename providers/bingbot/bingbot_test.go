package bingbot_test

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
		File("testdata/bingbot.json")

	ac := New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, IPv4Entry{netip.MustParsePrefix("40.77.139.0/25")})
	require.Contains(t, doc.IPv4Prefixes, IPv4Entry{netip.MustParsePrefix("20.15.133.160/27")})
	require.NotContains(t, doc.IPv4Prefixes, IPv4Entry{netip.MustParsePrefix("200.20.133.160/27")})
}
