package ovh

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
		File("testdata/prefixes.txt")

	o := New()
	gock.InterceptClient(o.Client.HTTPClient)

	prefixes, err := o.Fetch()
	require.NoError(t, err)
	require.Contains(t, prefixes, netip.MustParsePrefix("192.0.2.0/24"))
	require.Contains(t, prefixes, netip.MustParsePrefix("2001:db8::/32"))
}
