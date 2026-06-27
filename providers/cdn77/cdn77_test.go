package cdn77_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/cdn77"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(cdn77.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/prefixes.json")

	c := cdn77.New()
	gock.InterceptClient(c.Client.HTTPClient)

	doc, err := c.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 3)
	require.Len(t, doc.IPv6Prefixes, 2)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("185.59.220.0/22"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2a13:aac0::/29"))
}

func TestProcessDataWrappedObject(t *testing.T) {
	data := []byte(`{"prefixes":["185.59.220.0/22","2a13:aac0::/29"]}`)

	doc, err := cdn77.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 1)
	require.Len(t, doc.IPv6Prefixes, 1)
}
