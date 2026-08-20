package blocklistde_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/blocklistde"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(blocklistde.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/all.txt")

	p := blocklistde.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 40)
	require.Len(t, doc.IPv6Prefixes, 5)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("1.0.164.165/32"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/all.txt")
	require.NoError(t, err)

	doc, err := blocklistde.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 40)
	require.Len(t, doc.IPv6Prefixes, 5)
}

// bare addresses become host prefixes and comments are ignored.
func TestProcessDataNotation(t *testing.T) {
	doc, err := blocklistde.ProcessData([]byte("# comment\n1.2.3.4\n10.0.0.0/8\n\nnot-an-ip\n"))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("1.2.3.4/32"),
		netip.MustParsePrefix("10.0.0.0/8"),
	}, doc.IPv4Prefixes)
	require.Empty(t, doc.IPv6Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(blocklistde.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := blocklistde.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}

// the feed carries both address families.
func TestProcessDataIPv6(t *testing.T) {
	data, err := os.ReadFile("testdata/all.txt")
	require.NoError(t, err)

	doc, err := blocklistde.ProcessData(data)
	require.NoError(t, err)
	require.Contains(t, doc.IPv6Prefixes,
		netip.MustParsePrefix("2001:1670:26:f2e:7cc0:dc76:7bd7:bd68/128"))
}
