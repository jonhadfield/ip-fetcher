package cinsscore_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/cinsscore"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(cinsscore.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ci-badguys.txt")

	p := cinsscore.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 40)
	require.Empty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("1.119.158.77/32"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/ci-badguys.txt")
	require.NoError(t, err)

	doc, err := cinsscore.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 40)
	require.Empty(t, doc.IPv6Prefixes)
}

// bare addresses become host prefixes and comments are ignored.
func TestProcessDataNotation(t *testing.T) {
	doc, err := cinsscore.ProcessData([]byte("# comment\n1.2.3.4\n10.0.0.0/8\n\nnot-an-ip\n"))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("1.2.3.4/32"),
		netip.MustParsePrefix("10.0.0.0/8"),
	}, doc.IPv4Prefixes)
	require.Empty(t, doc.IPv6Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(cinsscore.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := cinsscore.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
