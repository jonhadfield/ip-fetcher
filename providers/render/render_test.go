package render_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/render"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

const TestASN = "397273"

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(fmt.Sprintf(render.DownloadURL, TestASN))
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/prefixes.json")

	ac := render.New()
	ac.ASNs = []string{TestASN}
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("216.24.56.0/22"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2600:1f16::/32"))
}
