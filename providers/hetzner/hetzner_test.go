package hetzner_test

import (
	"fmt"
	"net/netip"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

const TestASN = "24940"

func TestFetch(t *testing.T) {
	u, err := url.Parse(fmt.Sprintf(DownloadURL, TestASN))
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/prefixes.json")

	ac := New()
	ac.ASNs = []string{TestASN}
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("192.0.2.0/24"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2a01:4f8::/32"))
}
