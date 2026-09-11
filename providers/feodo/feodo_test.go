package feodo_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/feodo"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockFeodo(t *testing.T) *feodo.Feodo {
	t.Helper()

	u, err := url.Parse(feodo.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/feodo.txt")

	f := feodo.New()
	gock.InterceptClient(f.Client.HTTPClient)

	return &f
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	f := mockFeodo(t)

	doc, err := f.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 10)
	require.Empty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("185.117.90.6/32"))
}

// the list arrives behind a commented header, which must not reach the output.
func TestProcessDataIgnoresCommentHeader(t *testing.T) {
	data, err := os.ReadFile("testdata/feodo.txt")
	require.NoError(t, err)

	doc, err := feodo.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 10)

	for _, prefix := range doc.IPv4Prefixes {
		require.True(t, prefix.Addr().Is4())
		require.Equal(t, 32, prefix.Bits())
	}
}

// the list is IPv4 only today, but an IPv6 address must land in its own family.
func TestProcessDataSplitsFamilies(t *testing.T) {
	doc, err := feodo.ProcessData([]byte("# header\n5.34.178.161\n2001:db8::1\n"))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("5.34.178.161/32")}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2001:db8::1/128")}, doc.IPv6Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(feodo.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	f := feodo.New()
	gock.InterceptClient(f.Client.HTTPClient)

	_, err = f.Fetch()
	require.Error(t, err)
}
