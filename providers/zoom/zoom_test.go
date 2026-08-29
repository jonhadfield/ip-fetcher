package zoom_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/zoom"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(zoom.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/zoom.txt")

	p := zoom.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 25)
	require.Empty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("3.7.35.0/25"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/zoom.txt")
	require.NoError(t, err)

	doc, err := zoom.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 25)
	require.Empty(t, doc.IPv6Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(zoom.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := zoom.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
