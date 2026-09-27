package invicti_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/invicti"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockPages(t *testing.T) *invicti.Invicti {
	t.Helper()

	us, err := url.Parse(invicti.USDownloadURL)
	require.NoError(t, err)

	eu, err := url.Parse(invicti.EUDownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", us.Scheme, us.Host)).
		Get(us.Path).
		Reply(http.StatusOK).
		File("testdata/trustlist-us.html")

	gock.New(fmt.Sprintf("%s://%s", eu.Scheme, eu.Host)).
		Get(eu.Path).
		Reply(http.StatusOK).
		File("testdata/trustlist-eu.html")

	p := invicti.New()
	gock.InterceptClient(p.Client.HTTPClient)

	return &p
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	p := mockPages(t)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 8)
	require.Empty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("3.228.162.54/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("3.79.201.172/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("38.123.140.0/24"))
}

func TestFetchDataReturnsAddresses(t *testing.T) {
	defer gock.Off()

	p := mockPages(t)

	data, _, _, err := p.FetchData()
	require.NoError(t, err)
	require.NotContains(t, string(data), "<")
	require.Contains(t, string(data), "3.228.162.54")
	require.Contains(t, string(data), "3.79.201.172")
}

func TestFindAddressesDropsRepeatsAcrossRegions(t *testing.T) {
	us, err := os.ReadFile("testdata/trustlist-us.html")
	require.NoError(t, err)

	eu, err := os.ReadFile("testdata/trustlist-eu.html")
	require.NoError(t, err)

	addresses, err := invicti.FindAddresses(append(us, eu...))
	require.NoError(t, err)
	require.Equal(t, []string{
		"3.228.162.54",
		"52.0.216.190",
		"38.123.140.0/24",
		"54.85.4.50",
		"54.242.66.255",
		"3.79.201.172",
		"3.69.209.29",
		"18.194.203.224",
	}, addresses)
}

func TestFindAddressesMissing(t *testing.T) {
	_, err := invicti.FindAddresses([]byte("<html><body><td>no addresses</td></body></html>"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	us, err := url.Parse(invicti.USDownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", us.Scheme, us.Host)).
		Get(us.Path).
		Reply(http.StatusNotFound)

	p := invicti.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
