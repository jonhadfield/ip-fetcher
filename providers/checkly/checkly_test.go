package checkly_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/checkly"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockLists(t *testing.T) {
	t.Helper()

	for _, l := range []struct{ raw, file string }{
		{checkly.IPv4URL, "testdata/static-ips.json"},
		{checkly.IPv6URL, "testdata/static-ipv6s.json"},
	} {
		u, err := url.Parse(l.raw)
		require.NoError(t, err)

		gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
			Get(u.Path).
			Reply(http.StatusOK).
			File(l.file)
	}
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	mockLists(t)

	c := checkly.New()
	gock.InterceptClient(c.Client.HTTPClient)

	doc, err := c.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 15)
	require.Len(t, doc.IPv6Prefixes, 10)
}

// the two endpoints use different notation: bare addresses for IPv4, which
// become host prefixes, and real prefixes for IPv6.
func TestFetchHandlesBothNotations(t *testing.T) {
	defer gock.Off()

	mockLists(t)

	c := checkly.New()
	gock.InterceptClient(c.Client.HTTPClient)

	doc, err := c.Fetch()
	require.NoError(t, err)

	for _, p := range doc.IPv4Prefixes {
		require.Equal(t, 32, p.Bits(), "bare IPv4 addresses should become host prefixes")
	}

	for _, p := range doc.IPv6Prefixes {
		require.True(t, p.Addr().Is6())
		require.Less(t, p.Bits(), 128, "IPv6 endpoint publishes real prefixes")
	}
}

func TestProcessData(t *testing.T) {
	raw := `{"ipv4":["1.2.3.4"],"ipv6":["2001:db8::/56"]}`

	doc, err := checkly.ProcessData([]byte(raw))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("1.2.3.4/32")}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2001:db8::/56")}, doc.IPv6Prefixes)
}

func TestProcessDataSkipsInvalid(t *testing.T) {
	doc, err := checkly.ProcessData([]byte(`{"ipv4":["1.2.3.4","nonsense"],"ipv6":[]}`))
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 1)
	require.Empty(t, doc.IPv6Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(checkly.IPv4URL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	c := checkly.New()
	gock.InterceptClient(c.Client.HTTPClient)

	_, err = c.Fetch()
	require.Error(t, err)
}
