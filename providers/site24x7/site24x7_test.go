package site24x7_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/site24x7"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

// the export path the locations page fixture links to.
const exportPath = "/mesite24x7/location-manager/json/IP_Address_View/TESTTOKEN"

func mockLocations(t *testing.T) *site24x7.Site24x7 {
	t.Helper()

	u, err := url.Parse(site24x7.LocationsURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/locations.html")

	gock.New("https://creatorapp.zohopublic.in").
		Get(exportPath).
		Reply(http.StatusOK).
		File("testdata/locations.json")

	s := site24x7.New()
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
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("37.221.111.107/32"))
}

// most locations have no IPv6 address, and the empty field must not become a
// bogus prefix.
func TestFetchSkipsLocationsWithoutIPv6(t *testing.T) {
	defer gock.Off()

	s := mockLocations(t)

	doc, err := s.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv6Prefixes, 5)
	require.Less(t, len(doc.IPv6Prefixes), len(doc.IPv4Prefixes))
}

// FetchData returns the export rather than the page it was discovered on.
func TestFetchDataReturnsExport(t *testing.T) {
	defer gock.Off()

	s := mockLocations(t)

	data, _, _, err := s.FetchData()
	require.NoError(t, err)
	require.Contains(t, string(data), "IP_Address_View")

	doc, err := site24x7.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.Locations, 12)
}

// a few locations publish a prefix rather than a single address.
func TestProcessDataAcceptsPrefixes(t *testing.T) {
	data, err := os.ReadFile("testdata/locations.json")
	require.NoError(t, err)

	doc, err := site24x7.ProcessData(data)
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("185.230.212.0/23"))
}

// the export link is one of four flavours of the same view, and its href
// entities have to be decoded.
func TestFindExportURL(t *testing.T) {
	page, err := os.ReadFile("testdata/locations.html")
	require.NoError(t, err)

	exportURL, err := site24x7.FindExportURL(page)
	require.NoError(t, err)
	require.Equal(
		t,
		"https://creatorapp.zohopublic.in"+exportPath+"?src=content&format=json",
		exportURL,
	)
}

func TestFindExportURLMissing(t *testing.T) {
	_, err := site24x7.FindExportURL([]byte("<html><body>no links here</body></html>"))
	require.Error(t, err)
}

// a restructured locations page fails rather than returning the page itself.
func TestFetchWithoutExportLink(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(site24x7.LocationsURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		BodyString("<html><body>no links here</body></html>")

	s := site24x7.New()
	gock.InterceptClient(s.Client.HTTPClient)

	_, err = s.Fetch()
	require.Error(t, err)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := site24x7.ProcessData([]byte("not json"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(site24x7.LocationsURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	s := site24x7.New()
	gock.InterceptClient(s.Client.HTTPClient)

	_, err = s.Fetch()
	require.Error(t, err)
}

// several records carry trailing whitespace around their addresses.
func TestProcessDataTrimsWhitespace(t *testing.T) {
	raw := `{"IP_Address_View":[{"ID":"1","City":"Denver ","Place":"US","external_ip":"1.2.3.4 ","IPv6_Address_External":" 2001:db8::1"}]}`

	doc, err := site24x7.ProcessData([]byte(raw))
	require.NoError(t, err)
	require.Equal(t, "1.2.3.4", doc.Locations[0].IP)
	require.Equal(t, "Denver", doc.Locations[0].City)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("1.2.3.4/32")}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2001:db8::1/128")}, doc.IPv6Prefixes)
}
