package rapid7_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/rapid7"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockPage(t *testing.T) *rapid7.Rapid7 {
	t.Helper()

	u, err := url.Parse(rapid7.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/allowlist-cloud-engine-ips.html")

	p := rapid7.New()
	gock.InterceptClient(p.Client.HTTPClient)

	return &p
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	p := mockPage(t)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 9)
	require.Empty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("34.192.183.106/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("35.158.144.37/32"))
}

func TestFetchDataReturnsAddresses(t *testing.T) {
	defer gock.Off()

	p := mockPage(t)

	data, _, _, err := p.FetchData()
	require.NoError(t, err)
	require.NotContains(t, string(data), "<")
	require.Equal(t, "34.192.183.106", string(data)[:len("34.192.183.106")])
}

func TestFindAddressesIgnoresVersionNoise(t *testing.T) {
	page, err := os.ReadFile("testdata/allowlist-cloud-engine-ips.html")
	require.NoError(t, err)

	addresses, err := rapid7.FindAddresses(page)
	require.NoError(t, err)
	require.Len(t, addresses, 9)
	require.NotContains(t, addresses, "1.2.2.0")
	require.NotContains(t, addresses, "4.0.15.31")
}

func TestFindAddressesMissing(t *testing.T) {
	_, err := rapid7.FindAddresses([]byte("<html><body><td>US-1</td></body></html>"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(rapid7.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := rapid7.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
