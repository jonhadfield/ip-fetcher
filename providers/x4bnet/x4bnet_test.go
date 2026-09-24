package x4bnet_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/x4bnet"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u4, err := url.Parse(x4bnet.IPv4URL)
	require.NoError(t, err)
	u6, err := url.Parse(x4bnet.IPv6URL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u4.Scheme, u4.Host)).
		Get(u4.Path).
		Reply(http.StatusOK).
		File("testdata/ipv4.txt")
	gock.New(fmt.Sprintf("%s://%s", u6.Scheme, u6.Host)).
		Get(u6.Path).
		Reply(http.StatusOK).
		File("testdata/ipv6.txt")

	p := x4bnet.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("2.26.157.0/24"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("203.0.113.0/24"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2001:ac8::/32"))
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u4, err := url.Parse(x4bnet.IPv4URL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u4.Scheme, u4.Host)).
		Get(u4.Path).
		Reply(http.StatusNotFound)

	p := x4bnet.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
