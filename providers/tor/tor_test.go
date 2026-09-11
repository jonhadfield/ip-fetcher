package tor_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/tor"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockTor(t *testing.T) *tor.Tor {
	t.Helper()

	u, err := url.Parse(tor.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/tor.txt")

	t1 := tor.New()
	gock.InterceptClient(t1.Client.HTTPClient)

	return &t1
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	tr := mockTor(t)

	doc, err := tr.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 10)
	require.Len(t, doc.IPv6Prefixes, 2)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("185.220.101.34/32"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2a0b:f4c2:2::1/128"))
}

// the list holds bare addresses, so each must become a host prefix.
func TestProcessDataUsesHostPrefixes(t *testing.T) {
	data, err := os.ReadFile("testdata/tor.txt")
	require.NoError(t, err)

	doc, err := tor.ProcessData(data)
	require.NoError(t, err)

	for _, prefix := range doc.IPv4Prefixes {
		require.True(t, prefix.Addr().Is4())
		require.Equal(t, 32, prefix.Bits())
	}

	for _, prefix := range doc.IPv6Prefixes {
		require.True(t, prefix.Addr().Is6())
		require.Equal(t, 128, prefix.Bits())
	}
}

// a malformed entry must not discard the rest of the list.
func TestProcessDataSkipsInvalidEntry(t *testing.T) {
	doc, err := tor.ProcessData([]byte("185.220.101.34\nnot-an-ip\n\n2a0b:f4c2:2::1\n"))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("185.220.101.34/32")}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2a0b:f4c2:2::1/128")}, doc.IPv6Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(tor.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	tr := tor.New()
	gock.InterceptClient(tr.Client.HTTPClient)

	_, err = tr.Fetch()
	require.Error(t, err)
}
