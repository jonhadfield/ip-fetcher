package statuscake_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/statuscake"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockLocations(t *testing.T) *statuscake.StatusCake {
	t.Helper()

	u, err := url.Parse(statuscake.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		MatchParam("format", "json").
		Reply(http.StatusOK).
		File("testdata/locations.json")

	s := statuscake.New()
	gock.InterceptClient(s.Client.HTTPClient)

	return &s
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	s := mockLocations(t)

	doc, err := s.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.Locations, 12)
	require.Len(t, doc.IPv4Prefixes, 12)
}

// most locations have an IPv6 address but a good number do not, and the empty
// field must not become a bogus prefix.
func TestFetchSkipsLocationsWithoutIPv6(t *testing.T) {
	defer gock.Off()

	s := mockLocations(t)

	doc, err := s.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv6Prefixes, 8)
	require.Less(t, len(doc.IPv6Prefixes), len(doc.IPv4Prefixes))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/locations.json")
	require.NoError(t, err)

	doc, err := statuscake.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.Locations, 12)

	for _, p := range doc.IPv4Prefixes {
		require.True(t, p.Addr().Is4())
		require.Equal(t, 32, p.Bits())
	}

	for _, p := range doc.IPv6Prefixes {
		require.True(t, p.Addr().Is6())
		require.Equal(t, 128, p.Bits())
	}
}

// bare addresses become host prefixes and an empty ipv6 field is ignored.
func TestProcessDataNotation(t *testing.T) {
	raw := `{"1":{"ip":"1.2.3.4","ipv6":"2001:db8::1"},"2":{"ip":"5.6.7.8","ipv6":""}}`

	doc, err := statuscake.ProcessData([]byte(raw))
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 2)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2001:db8::1/128")}, doc.IPv6Prefixes)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := statuscake.ProcessData([]byte("not json"))
	require.Error(t, err)
}
