package url_test

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	mUrl "github.com/jonhadfield/ip-fetcher/providers/url"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestReadRawPrefixesFromFileData(t *testing.T) {
	d, err := os.ReadFile("testdata/ip-file-1.txt")
	require.NoError(t, err)
	require.NotEmpty(t, d)
	rp, err := mUrl.ReadRawPrefixesFromFileData(d)
	require.NoError(t, err)
	require.Len(t, rp, 4)
	require.Equal(t, "1.1.1.1/32", rp[0].String())
	require.Equal(t, "8.8.4.4/32", rp[1].String())
	require.Equal(t, "9.9.9.0/24", rp[3].String())
}

func TestFetchUrlData(t *testing.T) {
	u, err := url.Parse("https://www.example.com/files/ips.net")
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ip-file-1.txt")

	hf := mUrl.New()
	gock.InterceptClient(hf.HTTPClient.HTTPClient)

	response, err := mUrl.FetchURLResponse(hf.HTTPClient, "https://www.example.com/files/ips.net")
	require.NoError(t, err)
	require.NotEmpty(t, response.Data)
}

func TestFetchUrlsWithoutUrls(t *testing.T) {
	hf := mUrl.New()
	_, err := hf.Get([]mUrl.Request{})
	require.Error(t, err)
	require.ErrorContains(t, err, "no URLs to fetch")
}

func TestFetchUrls(t *testing.T) {
	u, err := url.Parse("https://www.example.com/files/ips.net")
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ip-file-1.txt")

	hf := mUrl.New()
	gock.InterceptClient(hf.HTTPClient.HTTPClient)
	responses, err := hf.Get([]mUrl.Request{
		{URL: u},
	})

	require.NoError(t, err)
	require.NotEmpty(t, responses)
}

func TestFetchUrlsWithFailedRequest(t *testing.T) {
	u, err := url.Parse("https://www.example.com/files/ips.net")
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusNotFound).
		File("testdata/ip-file-1.txt")

	hf := mUrl.New()
	gock.InterceptClient(hf.HTTPClient.HTTPClient)
	responses, err := hf.Get([]mUrl.Request{
		{
			URL: u,
		},
	})

	require.Error(t, err)
	require.Empty(t, responses)
}

// TestFetchPrefixesReportsWhyEveryFetchFailed covers a request that fails
// before any response arrives, which used to surface only as "no responses"
// with the cause, here a refused connection, logged at debug and dropped.
func TestFetchPrefixesReportsWhyEveryFetchFailed(t *testing.T) {
	u, err := url.Parse("http://127.0.0.1:1/files/ips.net")
	require.NoError(t, err)

	c := mUrl.New()
	c.HTTPClient.RetryMax = 0

	_, err = c.FetchPrefixes([]mUrl.Request{{URL: u}})
	require.ErrorContains(t, err, "no responses")
	require.ErrorContains(t, err, "failed to get http://127.0.0.1:1/files/ips.net")
	require.ErrorContains(t, err, "connection refused")

	_, err = c.FetchPrefixesAsText([]mUrl.Request{{URL: u}})
	require.ErrorContains(t, err, "connection refused")
}

func TestFetchPrefixesReportsStatus(t *testing.T) {
	defer gock.Off()

	gock.New("https://www.example.com").
		Get("/files/blocked.net").
		Reply(http.StatusForbidden)

	u, err := url.Parse("https://www.example.com/files/blocked.net")
	require.NoError(t, err)

	c := mUrl.New()
	gock.InterceptClient(c.HTTPClient.HTTPClient)

	_, err = c.FetchPrefixes([]mUrl.Request{{URL: u}})
	require.ErrorContains(t, err, "failed to get https://www.example.com/files/blocked.net: status 403")
}

// TestFetchPrefixesKeepsPartialResults checks that one failed URL does not
// discard the prefixes fetched from the others.
func TestFetchPrefixesKeepsPartialResults(t *testing.T) {
	defer gock.Off()

	gock.New("https://www.example.com").
		Get("/files/ips.net").
		Reply(http.StatusOK).
		File("testdata/ip-file-1.txt")
	gock.New("https://www.example.com").
		Get("/files/blocked.net").
		Reply(http.StatusForbidden)

	good, err := url.Parse("https://www.example.com/files/ips.net")
	require.NoError(t, err)
	bad, err := url.Parse("https://www.example.com/files/blocked.net")
	require.NoError(t, err)

	c := mUrl.New()
	gock.InterceptClient(c.HTTPClient.HTTPClient)

	prefixes, err := c.FetchPrefixes([]mUrl.Request{{URL: good}, {URL: bad}})
	require.NoError(t, err)
	require.Len(t, prefixes, 4)
}
