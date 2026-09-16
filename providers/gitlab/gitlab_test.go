package gitlab_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/gitlab"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(gitlab.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/gitlab.md")

	g := gitlab.New()
	gock.InterceptClient(g.Client.HTTPClient)

	doc, err := g.Fetch()
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("34.74.90.64/28"),
		netip.MustParsePrefix("34.74.226.0/24"),
	}, doc.IPv4Prefixes)
}

func TestFetchDataPublishesText(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(gitlab.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/gitlab.md")

	g := gitlab.New()
	gock.InterceptClient(g.Client.HTTPClient)

	data, _, status, err := g.FetchData()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "34.74.90.64/28\n34.74.226.0/24\n", string(data))
}

func TestFindAddressesMissing(t *testing.T) {
	_, err := gitlab.FindAddresses([]byte("no ranges here"))
	require.Error(t, err)
}
