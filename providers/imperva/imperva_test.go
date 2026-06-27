package imperva_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/imperva"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(imperva.DownloadURL)
	require.NoError(t, err)
	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Post(u.Path).
		Reply(http.StatusOK).
		File("testdata/ips.json")

	i := imperva.New()
	gock.InterceptClient(i.Client.HTTPClient)

	doc, err := i.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 3)
	require.Len(t, doc.IPv6Prefixes, 1)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("199.83.128.0/21"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2a02:e980::/29"))
}
