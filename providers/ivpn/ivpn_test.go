package ivpn_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/ivpn"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(ivpn.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/servers.json")

	p := ivpn.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("149.22.83.100/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("149.22.83.97/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("149.22.83.102/32"))
	require.NotContains(t, doc.IPv4Prefixes, netip.MustParsePrefix("172.16.0.1/12"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/servers.json")
	require.NoError(t, err)

	doc, err := ivpn.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 3)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(ivpn.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := ivpn.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
