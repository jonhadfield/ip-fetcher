package digitalocean_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jonhadfield/ip-fetcher/providers/digitalocean"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetchData(t *testing.T) {
	u, err := url.Parse(digitalocean.DigitaloceanDownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	lastModified := "Thu, 05 Jan 2023 19:43:47 GMT"
	etag := "63b72873-115c1"

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		SetHeader(web.ETagHeader, etag).
		SetHeader(web.LastModifiedHeader, lastModified).
		File("testdata/google.csv")

	ac := digitalocean.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	data, headers, status, err := ac.FetchData()
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.Len(t, headers.Values(web.ETagHeader), 1)
	require.Equal(t, etag, headers.Values(web.ETagHeader)[0])
	require.Len(t, headers.Values(web.LastModifiedHeader), 1)
	require.Equal(t, lastModified, headers.Values(web.LastModifiedHeader)[0])
	require.Equal(t, http.StatusOK, status)
}

func TestFetch(t *testing.T) {
	u, err := url.Parse(digitalocean.DigitaloceanDownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	lastModified := "Thu, 05 Jan 2023 19:43:47 GMT"
	etag := "63b72873-115c1"

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		SetHeader(web.ETagHeader, etag).
		SetHeader(web.LastModifiedHeader, lastModified).
		File("testdata/google.csv")

	ac := digitalocean.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.Records)
	require.Len(t, doc.Records, 1662)
	require.Equal(t, doc.ETag, etag)
	require.Equal(t, doc.LastModified.Format(time.RFC1123), lastModified)
}
