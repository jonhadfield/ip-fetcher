package gcore_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/gcore"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(gcore.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/public-ip-list.json")

	p := gcore.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 20)
	require.Len(t, doc.IPv6Prefixes, 15)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("80.15.248.3/32"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/public-ip-list.json")
	require.NoError(t, err)

	doc, err := gcore.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 20)
	require.Len(t, doc.IPv6Prefixes, 15)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(gcore.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := gcore.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
