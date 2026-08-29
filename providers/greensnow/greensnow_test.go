package greensnow_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/greensnow"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(greensnow.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/greensnow.txt")

	g := greensnow.New()
	gock.InterceptClient(g.Client.HTTPClient)

	doc, err := g.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 40)
	require.Empty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("79.124.56.146/32"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/greensnow.txt")
	require.NoError(t, err)

	doc, err := greensnow.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 40)
}

// bare addresses become host prefixes and unparseable entries are skipped.
func TestProcessDataNotation(t *testing.T) {
	doc, err := greensnow.ProcessData([]byte("# comment\n1.2.3.4\n10.0.0.0/8\n\nnot-an-ip\n2001:db8::1\n"))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("1.2.3.4/32"),
		netip.MustParsePrefix("10.0.0.0/8"),
	}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2001:db8::1/128")}, doc.IPv6Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(greensnow.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	g := greensnow.New()
	gock.InterceptClient(g.Client.HTTPClient)

	_, err = g.Fetch()
	require.Error(t, err)
}
