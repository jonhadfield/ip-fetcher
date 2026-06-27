package stripe_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/stripe"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	wh, err := url.Parse(stripe.WebhooksURL)
	require.NoError(t, err)
	gock.New(fmt.Sprintf("%s://%s", wh.Scheme, wh.Host)).
		Get(wh.Path).
		Reply(http.StatusOK).
		File("testdata/ips_webhooks.json")

	api, err := url.Parse(stripe.APIURL)
	require.NoError(t, err)
	gock.New(fmt.Sprintf("%s://%s", api.Scheme, api.Host)).
		Get(api.Path).
		Reply(http.StatusOK).
		File("testdata/ips_api.json")

	s := stripe.New()
	gock.InterceptClient(s.Client.HTTPClient)

	doc, err := s.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.Webhooks)
	require.NotEmpty(t, doc.API)
	// 3.18.12.63 is in both lists; the flattened union must dedupe it to a single /32.
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("3.18.12.63/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("35.154.171.200/32"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2600:1f18:2e2c:f400::/56"))
}
