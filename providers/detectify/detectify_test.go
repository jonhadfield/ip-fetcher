package detectify_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/detectify"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockPage(t *testing.T) *detectify.Detectify {
	t.Helper()

	u, err := url.Parse(detectify.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/scanner-ip-addresses.html")

	d := detectify.New()
	gock.InterceptClient(d.Client.HTTPClient)

	return &d
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	d := mockPage(t)

	doc, err := d.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 11)
	require.Empty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("52.17.98.131/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("3.7.173.162/32"))
}

// FetchData returns the addresses rather than the page they came from.
func TestFetchDataReturnsAddresses(t *testing.T) {
	defer gock.Off()

	d := mockPage(t)

	data, _, _, err := d.FetchData()
	require.NoError(t, err)
	require.NotContains(t, string(data), "<")
	require.Equal(t, "52.17.98.131", string(data)[:len("52.17.98.131")])
}

// the page's icons carry digits that look like an address, so only the marked
// up cells count, and the hostname cell is not one.
func TestFindAddressesIgnoresMarkupAndHostnames(t *testing.T) {
	page, err := os.ReadFile("testdata/scanner-ip-addresses.html")
	require.NoError(t, err)

	addresses, err := detectify.FindAddresses(page)
	require.NoError(t, err)
	require.Len(t, addresses, 11)
	require.NotContains(t, addresses, "scanner.detectify.com")
	require.NotContains(t, addresses, "138.112.25.25")
}

// an address repeated between sections is listed once.
func TestFindAddressesDropsRepeats(t *testing.T) {
	page := []byte(`<code>1.2.3.4</code><code>1.2.3.4</code><code>5.6.7.8</code>`)

	addresses, err := detectify.FindAddresses(page)
	require.NoError(t, err)
	require.Equal(t, []string{"1.2.3.4", "5.6.7.8"}, addresses)
}

// a restructured page fails rather than publishing an empty list.
func TestFindAddressesMissing(t *testing.T) {
	_, err := detectify.FindAddresses([]byte("<html><body>no addresses here</body></html>"))
	require.Error(t, err)
}

func TestFetchWithoutAddresses(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(detectify.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		BodyString("<html><body>no addresses here</body></html>")

	d := detectify.New()
	gock.InterceptClient(d.Client.HTTPClient)

	_, err = d.Fetch()
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(detectify.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	d := detectify.New()
	gock.InterceptClient(d.Client.HTTPClient)

	_, err = d.Fetch()
	require.Error(t, err)
}
