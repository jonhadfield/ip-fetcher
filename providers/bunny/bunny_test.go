package bunny_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/bunny"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockEndpoints(t *testing.T) {
	t.Helper()

	v4, err := url.Parse(bunny.IPv4URL)
	require.NoError(t, err)

	v6, err := url.Parse(bunny.IPv6URL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", v6.Scheme, v6.Host)).
		Get(v6.Path).
		Reply(http.StatusOK).
		File("testdata/edgeserverlist_ipv6.json")

	gock.New(fmt.Sprintf("%s://%s", v4.Scheme, v4.Host)).
		Get(v4.Path).
		Reply(http.StatusOK).
		File("testdata/edgeserverlist.json")
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	mockEndpoints(t)

	b := bunny.New()
	gock.InterceptClient(b.Client.HTTPClient)

	doc, err := b.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 4)
	require.Len(t, doc.IPv6Prefixes, 2)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("89.187.169.1/32"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2a0e:b107:540::1/128"))
}
