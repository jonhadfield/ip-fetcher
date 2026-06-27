package atlassian_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/atlassian"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(atlassian.DownloadURL)
	require.NoError(t, err)
	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ip-ranges.json")

	a := atlassian.New()
	gock.InterceptClient(a.Client.HTTPClient)

	doc, err := a.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.Items, 3)
	require.Len(t, doc.IPv4Prefixes, 2)
	require.Len(t, doc.IPv6Prefixes, 1)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("13.52.5.0/24"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2401:1d80::/32"))
}
