package m365_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/m365"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockM365(t *testing.T) *m365.M365 {
	t.Helper()

	u, err := url.Parse(m365.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/worldwide.json")

	m := m365.New()
	gock.InterceptClient(m.Client.HTTPClient)

	return &m
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	m := mockM365(t)

	doc, err := m.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.Endpoints, 3)
	require.Equal(t, "Exchange", doc.Endpoints[0].ServiceArea)
	require.Contains(t, doc.Endpoints[0].IPv4Prefixes, netip.MustParsePrefix("13.107.6.152/31"))
	require.Contains(t, doc.Endpoints[0].IPv6Prefixes, netip.MustParsePrefix("2603:1006::/40"))
}

// the category and required flag decide how an endpoint set is treated, so they
// must survive alongside the prefixes.
func TestFetchKeepsCategory(t *testing.T) {
	defer gock.Off()

	m := mockM365(t)

	doc, err := m.Fetch()
	require.NoError(t, err)

	teams := doc.Endpoints[2]
	require.Equal(t, "Microsoft Teams", teams.ServiceAreaDisplayName)
	require.Equal(t, "Optimize", teams.Category)
	require.True(t, teams.Required)
	require.True(t, teams.ExpressRoute)
	require.Equal(t, "3478,3479,3480,3481", teams.UDPPorts)
}

// endpoint sets published as URLs alone carry nothing to fetch.
func TestProcessDataDropsURLOnlySets(t *testing.T) {
	data, err := os.ReadFile("testdata/worldwide.json")
	require.NoError(t, err)

	doc, err := m365.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.Endpoints, 3)

	for _, endpoint := range doc.Endpoints {
		require.NotEmpty(t, append(endpoint.IPv4Prefixes, endpoint.IPv6Prefixes...))
		require.NotEqual(t, 46, endpoint.ID)
	}
}

func TestProcessDataSplitsFamilies(t *testing.T) {
	data, err := os.ReadFile("testdata/worldwide.json")
	require.NoError(t, err)

	doc, err := m365.ProcessData(data)
	require.NoError(t, err)

	for _, endpoint := range doc.Endpoints {
		for _, prefix := range endpoint.IPv4Prefixes {
			require.True(t, prefix.Addr().Is4())
		}

		for _, prefix := range endpoint.IPv6Prefixes {
			require.True(t, prefix.Addr().Is6())
		}
	}
}

// an unparseable address is skipped rather than discarding its endpoint set.
func TestProcessDataSkipsInvalidPrefix(t *testing.T) {
	doc, err := m365.ProcessData([]byte(`[{"id":1,"ips":["1.2.3.0/24","not-a-prefix"]}]`))
	require.NoError(t, err)
	require.Len(t, doc.Endpoints, 1)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("1.2.3.0/24")}, doc.Endpoints[0].IPv4Prefixes)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := m365.ProcessData([]byte("not json"))
	require.Error(t, err)
}

// the endpoints service will not answer without a caller GUID.
func TestDownloadURLCarriesClientRequestID(t *testing.T) {
	u, err := url.Parse(m365.DownloadURL)
	require.NoError(t, err)
	require.Equal(t, m365.DefaultClientRequestID, u.Query().Get("clientrequestid"))
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(m365.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	m := m365.New()
	gock.InterceptClient(m.Client.HTTPClient)

	_, err = m.Fetch()
	require.Error(t, err)
}
