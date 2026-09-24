package airvpn_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/airvpn"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(airvpn.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/status.json")

	p := airvpn.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("185.156.175.170/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("203.0.113.10/32"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2001:ac8:28:8::1/128"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/status.json")
	require.NoError(t, err)

	doc, err := airvpn.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 5)
	require.Len(t, doc.IPv6Prefixes, 4)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(airvpn.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := airvpn.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
