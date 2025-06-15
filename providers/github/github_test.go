package github_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/github"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(github.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/meta.json")

	gh := github.New()
	gock.InterceptClient(gh.Client.HTTPClient)

	prefixes, err := gh.Fetch()
	require.NoError(t, err)
	require.Contains(t, prefixes, netip.MustParsePrefix("192.30.252.0/22"))
	require.Contains(t, prefixes, netip.MustParsePrefix("140.82.112.0/20"))
}
