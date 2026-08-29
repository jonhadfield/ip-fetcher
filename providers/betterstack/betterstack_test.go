package betterstack_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/betterstack"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(betterstack.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ips.txt")

	b := betterstack.New()
	gock.InterceptClient(b.Client.HTTPClient)

	doc, err := b.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 17)
	require.Len(t, doc.IPv6Prefixes, 3)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("5.223.56.56/32"))
}

// the single list carries both families, so they must be split correctly.
func TestProcessDataSplitsFamilies(t *testing.T) {
	data, err := os.ReadFile("testdata/ips.txt")
	require.NoError(t, err)

	doc, err := betterstack.ProcessData(data)
	require.NoError(t, err)

	for _, p := range doc.IPv4Prefixes {
		require.True(t, p.Addr().Is4())
	}

	for _, p := range doc.IPv6Prefixes {
		require.True(t, p.Addr().Is6())
	}
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(betterstack.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	b := betterstack.New()
	gock.InterceptClient(b.Client.HTTPClient)

	_, err = b.Fetch()
	require.Error(t, err)
}
