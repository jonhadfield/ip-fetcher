package linode_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/linode"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetchData(t *testing.T) {
	u, err := url.Parse(linode.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	lastModified := "Thu, 05 Jan 2023 19:43:47 GMT"
	etag := "63b72873-115c1"

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		SetHeader(web.EtagHeader, etag).
		SetHeader(web.LastModifiedHeader, lastModified).
		File("testdata/prefixes.csv")

	ld := linode.New()
	defer gock.Off()

	gock.InterceptClient(ld.Client.HTTPClient)
	data, headers, status, err := ld.FetchData()
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.Len(t, headers.Values(web.EtagHeader), 1)
	require.Equal(t, etag, headers.Values(web.EtagHeader)[0])
	require.Len(t, headers.Values(web.LastModifiedHeader), 1)
	require.Equal(t, lastModified, headers.Values(web.LastModifiedHeader)[0])
	require.Equal(t, http.StatusOK, status)
	require.Len(t, data, 681)
}

func TestFetch(t *testing.T) {
	u, err := url.Parse(linode.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	lastModified := "Thu, 05 Jan 2023 19:43:47 GMT"
	etag := "63b72873-115c1"

	gock.New(urlBase).
		Get(u.Path).
		Times(2).
		Reply(http.StatusOK).
		SetHeader(web.EtagHeader, etag).
		SetHeader(web.LastModifiedHeader, lastModified).
		File("testdata/prefixes.csv")

	defer gock.Off()

	ld := linode.New()
	gock.InterceptClient(ld.Client.HTTPClient)

	data, headers, status, err := ld.FetchData()
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.Len(t, headers.Values(web.EtagHeader), 1)
	require.Equal(t, etag, headers.Values(web.EtagHeader)[0])
	require.Len(t, headers.Values(web.LastModifiedHeader), 1)
	require.Equal(t, lastModified, headers.Values(web.LastModifiedHeader)[0])
	require.Equal(t, http.StatusOK, status)
	require.Len(t, data, 681)

	doc, err := ld.Fetch()
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.NotEmpty(t, doc.Records)

	require.Equal(t, "US", doc.Records[0].Alpha2Code)
	require.Equal(t, "US-TX", doc.Records[0].Region)
	require.Equal(t, "Richardson", doc.Records[0].City)
	require.Equal(t, netip.MustParsePrefix("2600:3c00::/32"), doc.Records[0].Prefix)
}
