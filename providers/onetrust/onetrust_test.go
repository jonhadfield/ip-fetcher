package onetrust_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/onetrust"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(onetrust.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ips.txt")

	p := onetrust.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("20.54.106.120/29"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("20.212.124.24/29"))
	require.Empty(t, doc.IPv6Prefixes)
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/ips.txt")
	require.NoError(t, err)

	doc, err := onetrust.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 8)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(onetrust.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := onetrust.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
