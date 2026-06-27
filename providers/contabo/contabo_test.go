package contabo_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/contabo"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

const TestASN = "51167"

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(fmt.Sprintf(contabo.DownloadURL, TestASN))
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/prefixes.json")

	ac := contabo.New()
	ac.ASNs = []string{TestASN}
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("62.171.128.0/17"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2a02:c200::/32"))
}
