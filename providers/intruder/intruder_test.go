package intruder_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/intruder"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockPage(t *testing.T) *intruder.Intruder {
	t.Helper()

	u, err := url.Parse(intruder.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/allowlist.html")

	p := intruder.New()
	gock.InterceptClient(p.Client.HTTPClient)

	return &p
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	p := mockPage(t)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 3)
	require.Empty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("64.52.19.0/24"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("18.98.162.96/29"))
}

func TestFetchDataReturnsAddresses(t *testing.T) {
	defer gock.Off()

	p := mockPage(t)

	data, _, _, err := p.FetchData()
	require.NoError(t, err)
	require.NotContains(t, string(data), "<")
	require.Equal(t, "64.52.19.0/24", string(data)[:len("64.52.19.0/24")])
}

func TestFindAddressesDropsRepeats(t *testing.T) {
	page, err := os.ReadFile("testdata/allowlist.html")
	require.NoError(t, err)

	addresses, err := intruder.FindAddresses(page)
	require.NoError(t, err)
	require.Equal(t, []string{"64.52.19.0/24", "18.98.162.96/29", "13.115.104.128/25"}, addresses)
}

func TestFindAddressesMissing(t *testing.T) {
	_, err := intruder.FindAddresses([]byte("<html><body><code>not-an-address</code></body></html>"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(intruder.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := intruder.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
