package url

import (
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"gopkg.in/h2non/gock.v1"
)

func TestConstructor(t *testing.T) {
	hf := New()
	require.IsType(t, HttpFiles{}, hf)
	require.NotNil(t, hf.Client)

	hf.Add([]string{"https://www.example.com/files/ips.net"})
	require.Len(t, hf.Urls, 1)
	hf.Add([]string{"https://ww^w.example.com/files/ips.net", "https://www.example.com/files/ips2.txt"})
	require.Len(t, hf.Urls, 2)
}

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
	gock.InterceptClient(hf.Client.HTTPClient)

	response, err := fetchUrlResponse(hf.Client, "https://www.example.com/files/ips.net")
	require.NoError(t, err)
	require.NotEmpty(t, response.data)
}

func TestFetchUrlsWithoutUrls(t *testing.T) {
	hf := New()
	_, err := hf.FetchUrls()
	require.Error(t, err)
	require.ErrorContains(t, err, "no urls to fetch")
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
	hf.Add([]string{"https://www.example.com/files/ips.net"})
	gock.InterceptClient(hf.Client.HTTPClient)

	responses, err := hf.FetchUrls()
	require.NoError(t, err)
	require.NotEmpty(t, responses)
}

func TestFetchUrlsWithFailedRequest(t *testing.T) {
	u, err := url.Parse("https://www.example.com/files/ips.net")
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(404)

	hf := New()
	hf.Add([]string{"https://www.example.com/files/ips.net"})
	gock.InterceptClient(hf.Client.HTTPClient)

	_, err = hf.FetchUrls()
	require.Error(t, err)
}

func TestFetch(t *testing.T) {
	u, err := url.Parse("https://www.example.com/files/ips.net")
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/ip-file-1.txt")

	hf := New()
	hf.Add([]string{"https://www.example.com/files/ips.net"})
	gock.InterceptClient(hf.Client.HTTPClient)

	result, err := hf.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Contains(t, result, netip.MustParsePrefix("9.9.9.0/24"))
	require.Equal(t, result[netip.MustParsePrefix("9.9.9.0/24")], []string{"https://www.example.com/files/ips.net"})
	require.Equal(t, result[netip.MustParsePrefix("8.8.4.4/32")], []string{"https://www.example.com/files/ips.net"})
}

func TestFetchInvalidUrl(t *testing.T) {
	hf := New()
	hf.Add([]string{"https://ww^w.example.com/files/ips.net"})

	_, err := hf.Fetch()
	require.Error(t, err)
}

func TestFetchUrlsWithInvalidAndValidUrls(t *testing.T) {
	u, err := url.Parse("https://www.example.com/files/ips.net")
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/ip-file-1.txt")

	hf := New()
	hf.Add([]string{"https://ww^w.example.com/files/ips.net", "https://www.example.com/files/ips.net"})

	gock.InterceptClient(hf.Client.HTTPClient)

	result, err := hf.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Contains(t, result, netip.MustParsePrefix("9.9.9.0/24"))
	require.Equal(t, result[netip.MustParsePrefix("9.9.9.0/24")], []string{"https://www.example.com/files/ips.net"})
	require.Equal(t, result[netip.MustParsePrefix("8.8.4.4/32")], []string{"https://www.example.com/files/ips.net"})
}

func TestFetchUrlsWithInvalidAndValidUrlsWithInvalidPrefixes(t *testing.T) {
	u1, err := url.Parse("https://www.example.com/files/ips.net")
	require.NoError(t, err)
	urlBase1 := fmt.Sprintf("%s://%s", u1.Scheme, u1.Host)

	gock.New(urlBase1).
		Get(u1.Path).
		Reply(200).
		File("testdata/ip-file-1.txt")

	u2, err := url.Parse("https://www.example.com/files/extra/more.ips")
	require.NoError(t, err)
	urlBase2 := fmt.Sprintf("%s://%s", u2.Scheme, u2.Host)

	gock.New(urlBase2).
		Get(u2.Path).
		Reply(200).
		File("testdata/ip-file-2.txt")

	hf := New()
	hf.Add([]string{"https://ww^w.example.com/files/ips.net", "https://www.example.com/files/ips.net", "https://www.example.com/files/extra/more.ips"})

	gock.InterceptClient(hf.Client.HTTPClient)

	result, err := hf.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Contains(t, result, netip.MustParsePrefix("9.9.9.0/24"))
	require.Equal(t, result[netip.MustParsePrefix("9.9.9.0/24")], []string{"https://www.example.com/files/ips.net"})
	require.Equal(t, result[netip.MustParsePrefix("8.8.4.4/32")], []string{"https://www.example.com/files/ips.net"})
	require.Equal(t, result[netip.MustParsePrefix("1.62.4.25/32")], []string{"https://www.example.com/files/extra/more.ips"})
}

// func (hf *HttpFiles) FetchPrefixesAsText() (prefixes []string, err error) {
//	urlResponses, err := hf.FetchUrls()
//	if err != nil {
//		return
//	}
//
//	pum, err := GetPrefixURLMapFromUrlResponses(urlResponses)
//
//	for k := range pum {
//		prefixes = append(prefixes, k.String())
//	}
//
//	return
// }

func TestFetchPrefixesAsText(t *testing.T) {
	u, err := url.Parse("https://www.example.com/files/ips.net")
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/ip-file-1.txt")

	hf := New()
	hf.Add([]string{"https://www.example.com/files/ips.net"})
	gock.InterceptClient(hf.Client.HTTPClient)

	result, err := hf.FetchPrefixesAsText()
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.True(t, slices.Contains(result, "9.9.9.0/24"))
	require.True(t, slices.Contains(result, "8.8.4.4/32"))
}
