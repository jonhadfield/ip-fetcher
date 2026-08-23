package duckduckbot_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jonhadfield/ip-fetcher/providers/duckduckbot"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(duckduckbot.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/duckduckbot.json")

	ac := duckduckbot.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Contains(t, doc.IPv4Prefixes, duckduckbot.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("104.43.54.127/32")})
	require.Contains(t, doc.IPv4Prefixes, duckduckbot.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("13.86.35.212/32")})
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/duckduckbot.json")
	require.NoError(t, err)

	doc, err := duckduckbot.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 4)
	require.Empty(t, doc.IPv6Prefixes)
	require.Contains(t, doc.IPv4Prefixes, duckduckbot.IPv4Entry{IPv4Prefix: netip.MustParsePrefix("128.203.132.152/32")})
	require.Equal(t, time.Date(2026, time.July, 3, 15, 15, 37, 0, time.UTC), doc.CreationTime)
}
