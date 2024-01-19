package url

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestReadRawPrefixesFromFileData(t *testing.T) {
	d, err := os.ReadFile("testdata/ip-file-1.txt")
	require.NoError(t, err)
	require.NotEmpty(t, d)
	rp, err := ReadRawPrefixesFromFileData(d)
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
		Reply(200).
		File("testdata/ip-file-1.txt")

	hf := New()
	gock.InterceptClient(hf.HttpClient.HTTPClient)

	response, err := fetchUrlResponse(hf.HttpClient, "https://www.example.com/files/ips.net")
	require.NoError(t, err)
	require.NotEmpty(t, response.data)
}

func TestFetchUrlsWithoutUrls(t *testing.T) {
	hf := New()
	_, err := hf.Get([]Request{})
	require.Error(t, err)
	require.ErrorContains(t, err, "no URLs to fetch")
}

func TestFetchUrls(t *testing.T) {
	u, err := url.Parse("https://www.example.com/files/ips.net")
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/ip-file-1.txt")

	hf := New()
	gock.InterceptClient(hf.HttpClient.HTTPClient)
	responses, err := hf.Get([]Request{
		{Url: u},
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

	hf := New()
	gock.InterceptClient(hf.HttpClient.HTTPClient)
	responses, err := hf.Get([]Request{
		{
			Url: u,
		}})

	require.Error(t, err)
	require.Empty(t, responses)
}
