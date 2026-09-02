package sentry_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/sentry"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockSentry(t *testing.T) *sentry.Sentry {
	t.Helper()

	u, err := url.Parse(sentry.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/sentry.txt")

	s := sentry.New()
	gock.InterceptClient(s.Client.HTTPClient)

	return &s
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	s := mockSentry(t)

	doc, err := s.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 12)
	require.Empty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("34.123.33.225/32"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/sentry.txt")
	require.NoError(t, err)

	doc, err := sentry.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 12)

	for _, p := range doc.IPv4Prefixes {
		require.True(t, p.Addr().Is4())
		require.Equal(t, 32, p.Bits())
	}
}

// the list is IPv4 only today, but an IPv6 address must land in its own family,
// and a malformed entry must not discard the rest of the list.
func TestProcessDataMixedNotation(t *testing.T) {
	doc, err := sentry.ProcessData([]byte("34.123.33.225\n\nnot-an-ip\n10.0.0.0/8\n2001:db8::1\n"))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("34.123.33.225/32"),
		netip.MustParsePrefix("10.0.0.0/8"),
	}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2001:db8::1/128")}, doc.IPv6Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(sentry.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	s := sentry.New()
	gock.InterceptClient(s.Client.HTTPClient)

	_, err = s.Fetch()
	require.Error(t, err)
}
