package icloudpr_test

import (
	"fmt"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"net/netip"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetchData(t *testing.T) {
	u, err := url.Parse(DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	lastModified := "Thu, 05 Jan 2023 19:43:47 GMT"
	etag := "63b72873-115c1"

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		SetHeader("etag", etag).
		SetHeader(web.LastModifiedHeader, lastModified).
		File("testdata/egress-ip-ranges.csv")

	ld := New()
	defer gock.Off()

	gock.InterceptClient(ld.Client.HTTPClient)
	data, headers, status, err := ld.FetchData()
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.Len(t, headers.Values("etag"), 1)
	require.Equal(t, etag, headers.Values("etag")[0])
	require.Len(t, headers.Values(web.LastModifiedHeader), 1)
	require.Equal(t, lastModified, headers.Values(web.LastModifiedHeader)[0])
	require.Equal(t, 200, status)
	require.Len(t, data, 323)
}

func TestFetch(t *testing.T) {
	u, err := url.Parse(DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	lastModified := "Thu, 05 Jan 2023 19:43:47 GMT"
	etag := "662f6d82-b01feb"

	gock.New(urlBase).
		Get(u.Path).
		Times(2).
		Reply(200).
		SetHeader("etag", etag).
		SetHeader(web.LastModifiedHeader, lastModified).
		File("testdata/egress-ip-ranges.csv")

	defer gock.Off()

	ld := New()
	gock.InterceptClient(ld.Client.HTTPClient)

	data, headers, status, err := ld.FetchData()
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.Len(t, headers.Values("etag"), 1)
	require.Equal(t, etag, headers.Values("etag")[0])
	require.Len(t, headers.Values(web.LastModifiedHeader), 1)
	require.Equal(t, lastModified, headers.Values(web.LastModifiedHeader)[0])
	require.Equal(t, 200, status)
	require.Len(t, data, 323)

	doc, err := ld.Fetch()
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.NotEmpty(t, doc.Records)

	require.Equal(t, "GB", doc.Records[0].Alpha2Code)
	require.Equal(t, "GB-EN", doc.Records[0].Region)
	require.Equal(t, "London", doc.Records[0].City)
	require.Equal(t, netip.MustParsePrefix("172.224.224.0/27"), doc.Records[0].Prefix)
}
