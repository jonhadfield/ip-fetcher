package xpanse_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/xpanse"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockPage(t *testing.T) *xpanse.Xpanse {
	t.Helper()

	u, err := url.Parse(xpanse.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/scanning-activity.html")

	p := xpanse.New()
	gock.InterceptClient(p.Client.HTTPClient)

	return &p
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	p := mockPage(t)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 3)
	require.Len(t, doc.IPv6Prefixes, 2)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("35.203.210.0/23"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2604:a940:300:5b6::/64"))
}

func TestFetchDataReturnsAddresses(t *testing.T) {
	defer gock.Off()

	p := mockPage(t)

	data, _, _, err := p.FetchData()
	require.NoError(t, err)
	require.NotContains(t, string(data), "<")
	require.Equal(t, "35.203.210.0/23", string(data)[:len("35.203.210.0/23")])
}

func TestFindAddressesIgnoresBareAddresses(t *testing.T) {
	page, err := os.ReadFile("testdata/scanning-activity.html")
	require.NoError(t, err)

	addresses, err := xpanse.FindAddresses(page)
	require.NoError(t, err)
	require.Len(t, addresses, 5)
	require.NotContains(t, addresses, "1.2.3.4")
}

func TestFindAddressesMissing(t *testing.T) {
	_, err := xpanse.FindAddresses([]byte("<html><body>1.2.3.4</body></html>"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(xpanse.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := xpanse.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
