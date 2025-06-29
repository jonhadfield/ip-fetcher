package azure_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/azure"

	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/stretchr/testify/require"

	"gopkg.in/h2non/gock.v1"
)

const (
	// testDownloadURL     = "https://download.microsoft.com/download/7/1/D/71D86715-5596-4529-9B13-DA13A5DE5B63/ServiceTags_Public_20221212.json"
	testDownloadURL     = azure.WorkaroundDownloadURL
	testInitialURL      = "https://www.microsoft.com/en-us/download/confirmation.aspx?id=00000"
	testInitialFilePath = "testdata/initial.html"
	testDataFilePath    = "testdata/ServiceTags_Public_20221212.json"
)

// func TestGetDownloadURL(t *testing.T) {
// 	defer gock.Off()
//
// 	u, err := url.Parse(testInitialURL)
// 	require.NoError(t, err)
//
// 	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
// 	gock.New(urlBase).
// 		MatchParam("id", "00000").
// 		Get(u.Path).
// 		Reply(http.StatusOK).
// 		File(testInitialFilePath)
//
// 	ac := New()
// 	ac.InitialURL = testInitialURL
// 	gock.InterceptClient(ac.Client.HTTPClient)
//
// 	dURL, err := ac.GetDownloadURL()
// 	require.NoError(t, err)
// 	require.NotEmpty(t, dURL)
// 	require.Equal(t, "https://download.microsoft.com/download/7/1/D/71D86715-5596-4529-9B13-DA13A5DE5B63/ServiceTags_Public_2000000.json", dURL)
// }

func TestFetchRaw(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(testDownloadURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	exMD5 := "0Pl1673GWSGnCAHlQJ5pXA=="
	exEtag := "0x8DADCD65EF6DD96"
	exTimeStamp := "Tue, 13 Dec 2022 06:50:50 GMT"
	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		AddHeader(web.LastModifiedHeader, exTimeStamp).
		AddHeader(web.ContentMD5Header, exMD5).
		AddHeader(web.ETagHeader, exEtag).
		File(testDataFilePath)

	ac := azure.New()
	ac.DownloadURL = testDownloadURL
	gock.InterceptClient(ac.Client.HTTPClient)

	data, header, status, err := ac.FetchData()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, exMD5, header.Get(web.ContentMD5Header))
	require.Len(t, data, 2938956)
}

func TestFetchRawNoDownloadURL(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(testInitialURL)
	require.NoError(t, err)

	// intercept initial url
	gock.New(testInitialURL).
		Get(u.Path).
		Reply(http.StatusNotFound)

	_, err = url.Parse(testInitialURL)

	require.NoError(t, err)

	ac := azure.New()
	ac.InitialURL = testInitialURL
	gock.InterceptClient(ac.Client.HTTPClient)

	_, _, _, err = ac.FetchData()
	require.Error(t, err)
}

func TestFetchRawFailure(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(testDownloadURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusNotFound).
		File(testDataFilePath)

	ac := azure.New()
	ac.DownloadURL = testDownloadURL
	gock.InterceptClient(ac.Client.HTTPClient)

	data, _, status, err := ac.FetchData()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, status)
	require.Empty(t, data)
}

// func TestGetDownloadURLFailure(t *testing.T) {
// 	defer gock.Off()
//
// 	t.Parallel()
//
// 	u, err := url.Parse(testInitialURL)
// 	require.NoError(t, err)
//
// 	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
// 	exMD5 := "0Pl1674GWSGnCAHlQJ5pXA=="
// 	exEtag := "0x8DADCD65EF6DD96"
// 	exTimeStamp := "Tue, 13 Dec 2022 06:50:50 GMT"
// 	gock.New(urlBase).
// 		MatchParam("id", "00000").
// 		Get(u.Path).
// 		Reply(404).
// 		AddHeader("Last-Modified", exTimeStamp).
// 		AddHeader("Content-MD5", exMD5).
// 		AddHeader("ETag", exEtag)
//
// 	ac := New()
// 	ac.InitialURL = testInitialURL
// 	gock.InterceptClient(ac.Client.HTTPClient)
//
// 	_, err = ac.GetDownloadURL()
// 	require.Error(t, err)
// 	require.Contains(t, err.Error(), errFailedToDownload)
// }

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(testDownloadURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	exMD5 := "0Pl1674GWSGnCAHlQJ5pXA=="
	exEtag := "0x8DADCD65EF6DD96"
	exTimeStamp := "Tue, 13 Dec 2022 06:50:50 GMT"
	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		AddHeader(web.LastModifiedHeader, exTimeStamp).
		AddHeader(web.ContentMD5Header, exMD5).
		AddHeader(web.ETagHeader, exEtag).
		File(testDataFilePath)

	ac := azure.New()
	ac.DownloadURL = testDownloadURL
	gock.InterceptClient(ac.Client.HTTPClient)

	prefixes, _, err := ac.Fetch()
	require.NoError(t, err)

	ac.DownloadURL = urlBase

	require.Equal(t, "Public", prefixes.Cloud)
	require.Equal(t, 232, prefixes.ChangeNumber)
	require.Len(t, prefixes.Values, 2643)
}
