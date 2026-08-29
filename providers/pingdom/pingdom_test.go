package pingdom_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/pingdom"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockLists(t *testing.T) {
	t.Helper()

	for _, l := range []struct{ raw, file string }{
		{pingdom.IPv4URL, "testdata/probes-ipv4.txt"},
		{pingdom.IPv6URL, "testdata/probes-ipv6.txt"},
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

	p := pingdom.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 20)
	require.Len(t, doc.IPv6Prefixes, 12)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("3.10.222.182/32"))
}

// the probe lists are bare addresses, which become host prefixes.
func TestFetchYieldsHostPrefixes(t *testing.T) {
	defer gock.Off()

	mockLists(t)

	p := pingdom.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)

	for _, prefix := range doc.IPv4Prefixes {
		require.Equal(t, 32, prefix.Bits())
	}

	for _, prefix := range doc.IPv6Prefixes {
		require.Equal(t, 128, prefix.Bits())
	}
}

func TestProcessData(t *testing.T) {
	raw := `{"ipv4":["1.2.3.4/32"],"ipv6":["2001:db8::1/128"]}`

	doc, err := pingdom.ProcessData([]byte(raw))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("1.2.3.4/32")}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2001:db8::1/128")}, doc.IPv6Prefixes)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := pingdom.ProcessData([]byte("not json"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(pingdom.IPv4URL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := pingdom.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
