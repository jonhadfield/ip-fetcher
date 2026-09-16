package cachefly_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/cachefly"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(cachefly.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/cachefly.txt")

	c := cachefly.New()
	gock.InterceptClient(c.Client.HTTPClient)

	doc, err := c.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("205.234.175.0/24"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/cachefly.txt")
	require.NoError(t, err)

	doc, err := cachefly.ProcessData(data)
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(cachefly.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	c := cachefly.New()
	gock.InterceptClient(c.Client.HTTPClient)

	_, err = c.Fetch()
	require.Error(t, err)
}
