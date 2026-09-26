package nodeping_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/nodeping"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(nodeping.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/pinghosts.txt")

	p := nodeping.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("104.247.192.170/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("38.114.123.177/32"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2607:3f00:11:21::10/128"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/pinghosts.txt")
	require.NoError(t, err)

	doc, err := nodeping.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 2)
	require.Len(t, doc.IPv6Prefixes, 2)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(nodeping.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusForbidden)

	p := nodeping.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
