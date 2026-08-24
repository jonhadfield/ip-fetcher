package ahrefs_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/ahrefs"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(ahrefs.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ahrefs.json")

	ac := ahrefs.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, ahrefs.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("54.36.148.0/23")})
	require.Empty(t, doc.IPv6Prefixes)
	require.True(t, doc.CreationTime.IsZero())
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/ahrefs.json")
	require.NoError(t, err)

	doc, err := ahrefs.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 4)
	require.Contains(t, doc.IPv4Prefixes, ahrefs.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("198.244.240.0/24")})
	require.Empty(t, doc.IPv6Prefixes)
	require.True(t, doc.CreationTime.IsZero())
}

func TestProcessDataWithCreationTime(t *testing.T) {
	data := []byte(`{"creationTime":"2026-01-02T03:04:05.000000","prefixes":[{"ipv4Prefix":"5.39.1.224/27"},{"ipv6Prefix":"2001:db8::/32"}]}`)

	doc, err := ahrefs.ProcessData(data)
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, ahrefs.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("5.39.1.224/27")})
	require.Contains(t, doc.IPv6Prefixes, ahrefs.IPv6Entry{IPv6Prefix: netip.MustParsePrefix("2001:db8::/32")})
	require.False(t, doc.CreationTime.IsZero())
	require.Equal(t, 2026, doc.CreationTime.Year())
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := ahrefs.ProcessData([]byte("not json"))
	require.Error(t, err)
}
