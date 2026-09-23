package ipsum_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/ipsum"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(ipsum.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ipsum.txt")

	ip := ipsum.New()
	gock.InterceptClient(ip.Client.HTTPClient)

	doc, err := ip.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("94.154.43.254/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("192.0.2.1/32"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/ipsum.txt")
	require.NoError(t, err)

	doc, err := ipsum.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 4)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(ipsum.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	ip := ipsum.New()
	gock.InterceptClient(ip.Client.HTTPClient)

	_, err = ip.Fetch()
	require.Error(t, err)
}
