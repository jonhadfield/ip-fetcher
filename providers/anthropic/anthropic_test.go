package anthropic_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/anthropic"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(anthropic.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/anthropic.json")

	ac := anthropic.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, anthropic.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("216.73.216.0/22")})
	require.Contains(t, doc.IPv4Prefixes, anthropic.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("34.162.230.222/32")})
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/anthropic.json")
	require.NoError(t, err)

	doc, err := anthropic.ProcessData(data)
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, anthropic.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("35.245.175.129/32")})
	require.Equal(t, 2026, doc.CreationTime.Year())
}

// the sibling crawler feeds publish fractional seconds rather than RFC3339, so
// both must be accepted.
func TestProcessDataAlternativeTimeFormat(t *testing.T) {
	doc, err := anthropic.ProcessData([]byte(
		`{"creationTime":"2026-08-12T01:43:14.123456","prefixes":[{"ipv6Prefix":"2600:1900::/28"}]}`))
	require.NoError(t, err)
	require.Equal(t, 2026, doc.CreationTime.Year())
	require.Empty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv6Prefixes, anthropic.IPv6Entry{IPv6Prefix: netip.MustParsePrefix("2600:1900::/28")})
}

// creationTime is optional, so its absence must not fail the parse.
func TestProcessDataMissingCreationTime(t *testing.T) {
	doc, err := anthropic.ProcessData([]byte(`{"prefixes":[{"ipv4Prefix":"216.73.216.0/22"}]}`))
	require.NoError(t, err)
	require.True(t, doc.CreationTime.IsZero())
	require.Len(t, doc.IPv4Prefixes, 1)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := anthropic.ProcessData([]byte(`not json`))
	require.Error(t, err)
}
