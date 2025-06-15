package ovh_test

import (
	"fmt"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/ovh"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	cURL := fmt.Sprintf(ovh.DownloadURL, ovh.ASNs[0])
	u, err := url.Parse(cURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/prefixes.json")

	o := ovh.New()
	gock.InterceptClient(o.Client.HTTPClient)

	doc, err := o.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("192.0.2.0/24"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2001:db8::/32"))
}
