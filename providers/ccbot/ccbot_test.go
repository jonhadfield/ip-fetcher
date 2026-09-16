package ccbot_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/ccbot"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(ccbot.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ccbot.json")

	c := ccbot.New()
	gock.InterceptClient(c.Client.HTTPClient)

	doc, err := c.Fetch()
	require.NoError(t, err)
	require.NotZero(t, doc.CreationTime)
	require.Contains(t, doc.IPv4Prefixes, ccbot.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("3.41.188.32/29")})
	require.Contains(t, doc.IPv6Prefixes, ccbot.IPv6Entry{IPv6Prefix: netip.MustParsePrefix("2600:1f28:365:8000::/56")})
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/ccbot.json")
	require.NoError(t, err)

	doc, err := ccbot.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 4)
	require.Len(t, doc.IPv6Prefixes, 1)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(ccbot.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	c := ccbot.New()
	gock.InterceptClient(c.Client.HTTPClient)

	_, err = c.Fetch()
	require.Error(t, err)
}
