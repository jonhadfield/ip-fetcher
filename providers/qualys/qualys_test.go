package qualys_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/qualys"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

const testASN = "27385"

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(fmt.Sprintf(qualys.DownloadURL, testASN))
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/prefixes.json")

	p := qualys.New()
	p.ASNs = []string{testASN}
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("64.39.96.0/24"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("69.67.181.0/24"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2602:fdaa:c1::/48"))
}
