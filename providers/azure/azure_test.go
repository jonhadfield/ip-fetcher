package azure_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

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

// stubPage returns a PageFetcher serving fixed content, standing in for the
// cycletls call so discovery can be tested without reaching the live page.
func stubPage(body string, status int, err error) azure.PageFetcher {
	return func(string) (string, int, error) {
		return body, status, err
	}
}

func TestGetDownloadURL(t *testing.T) {
	page, err := os.ReadFile(testInitialFilePath)
	require.NoError(t, err)

	ac := azure.New()
	ac.InitialURL = testInitialURL
	ac.PageFetcher = stubPage(string(page), http.StatusOK, nil)

	dURL, err := ac.GetDownloadURL()
	require.NoError(t, err)
	require.Equal(t,
		"https://download.microsoft.com/download/7/1/D/71D86715-5596-4529-9B13-DA13A5DE5B63/ServiceTags_Public_2000000.json",
		dURL)
}

// a failing fetch is reported rather than silently yielding an empty URL.
func TestGetDownloadURLFetchError(t *testing.T) {
	ac := azure.New()
	ac.InitialURL = testInitialURL
	ac.PageFetcher = stubPage("", 0, errors.New("boom"))

	_, err := ac.GetDownloadURL()
	require.Error(t, err)
}

func TestGetDownloadURLBadStatus(t *testing.T) {
	ac := azure.New()
	ac.InitialURL = testInitialURL
	ac.PageFetcher = stubPage("", http.StatusNotFound, nil)

	_, err := ac.GetDownloadURL()
	require.Error(t, err)
}

// a page with no download link yields an empty URL, which FetchData treats as
// a discovery failure and falls back on.
func TestParseDownloadURLNoLink(t *testing.T) {
	require.Empty(t, azure.ParseDownloadURL(`<a href="https://example.com/nope">x</a>`))
	require.Empty(t, azure.ParseDownloadURL(""))
}

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

// with no DownloadURL set, FetchData runs discovery. The page fetch is stubbed
// as failing so it falls back to WorkaroundDownloadURL, which gock can serve.
func TestFetchRawNoDownloadURL(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(azure.WorkaroundDownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File(testDataFilePath)

	ac := azure.New()
	ac.InitialURL = testInitialURL
	ac.PageFetcher = stubPage("", http.StatusNotFound, nil)
	gock.InterceptClient(ac.Client.HTTPClient)

	data, _, status, err := ac.FetchData()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, data)
	require.Equal(t, azure.WorkaroundDownloadURL, ac.DownloadURL)
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

// Snapshots are published on Mondays, and the previous week's file is still
// served for a while, so the newest candidate must be tried first.
func TestCandidateDatesAreMondaysNewestFirst(t *testing.T) {
	// a Saturday
	dates := azure.CandidateDates(time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC))
	require.Len(t, dates, 3)
	require.Equal(t, "2026-08-24", dates[0].Format("2006-01-02"))
	require.Equal(t, "2026-08-17", dates[1].Format("2006-01-02"))
	require.Equal(t, "2026-08-10", dates[2].Format("2006-01-02"))

	for _, d := range dates {
		require.Equal(t, time.Monday, d.Weekday())
	}
}

// on a Monday the search starts that same day, not the week before.
func TestCandidateDatesOnMonday(t *testing.T) {
	dates := azure.CandidateDates(time.Date(2026, time.August, 24, 9, 0, 0, 0, time.UTC))
	require.Equal(t, "2026-08-24", dates[0].Format("2006-01-02"))
}

func TestFindDownloadURLTakesTheNewestHit(t *testing.T) {
	defer gock.Off()

	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)

	// newest missing, previous week present: the previous week must win, and
	// only after the newer one has been ruled out.
	newest := fmt.Sprintf(azure.DatedURLTemplate, "20260824")
	previous := fmt.Sprintf(azure.DatedURLTemplate, "20260817")

	for raw, code := range map[string]int{newest: http.StatusNotFound, previous: http.StatusOK} {
		u, err := url.Parse(raw)
		require.NoError(t, err)

		gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
			Head(u.Path).
			Reply(code)
	}

	ac := azure.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	require.Equal(t, previous, ac.FindDownloadURL(now))
}

// with nothing published, the caller is told so and can fall back.
func TestFindDownloadURLNoneFound(t *testing.T) {
	defer gock.Off()

	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)

	for _, d := range []string{"20260824", "20260817", "20260810"} {
		u, err := url.Parse(fmt.Sprintf(azure.DatedURLTemplate, d))
		require.NoError(t, err)

		gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
			Head(u.Path).
			Reply(http.StatusNotFound)
	}

	ac := azure.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	require.Empty(t, ac.FindDownloadURL(now))
}
