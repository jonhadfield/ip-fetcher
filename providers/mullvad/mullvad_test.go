package mullvad_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/mullvad"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(mullvad.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/mullvad.json")

	m := mullvad.New()
	gock.InterceptClient(m.Client.HTTPClient)

	doc, err := m.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("103.124.165.2/32"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2a04:27c0:0:e::f001/128"))
}

func TestProcessDataSkipsInactive(t *testing.T) {
	doc, err := mullvad.ProcessData([]byte(`[
		{"hostname":"a","active":false,"ipv4_addr_in":"1.2.3.4","ipv6_addr_in":""},
		{"hostname":"b","active":true,"ipv4_addr_in":"5.6.7.8","ipv6_addr_in":"2001:db8::1"}
	]`))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("5.6.7.8/32")}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2001:db8::1/128")}, doc.IPv6Prefixes)
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/mullvad.json")
	require.NoError(t, err)

	doc, err := mullvad.ProcessData(data)
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(mullvad.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	m := mullvad.New()
	gock.InterceptClient(m.Client.HTTPClient)

	_, err = m.Fetch()
	require.Error(t, err)
}
