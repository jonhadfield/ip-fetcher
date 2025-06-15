package scaleway_test

import (
	"fmt"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/scaleway"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(fmt.Sprintf(scaleway.DownloadURL, scaleway.ASNs[0]))
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/prefixes.json")

	o := scaleway.New()
	gock.InterceptClient(o.Client.HTTPClient)

	doc, err := o.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("51.15.0.0/16"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2001:bc8::/32"))
}
