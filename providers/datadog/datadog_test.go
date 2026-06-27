package datadog_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/datadog"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(datadog.DownloadURL)
	require.NoError(t, err)
	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ip-ranges.json")

	d := datadog.New()
	gock.InterceptClient(d.Client.HTTPClient)

	doc, err := d.Fetch()
	require.NoError(t, err)
	require.Equal(t, 61, doc.Version)
	require.Contains(t, doc.Categories, "agents")
	require.Contains(t, doc.Categories, "webhooks")
	// 3.233.144.0/20 appears in both agents and api; the flattened union must dedupe it.
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("3.233.144.0/20"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("44.207.0.0/16"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2600:1f18::/32"))
	require.Len(t, doc.IPv4Prefixes, 4)
}
