package circleci_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/circleci"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(circleci.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/circleci.json")

	c := circleci.New()
	gock.InterceptClient(c.Client.HTTPClient)

	doc, err := c.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.MacOS, netip.MustParsePrefix("100.27.248.128/25"))
	require.Contains(t, doc.Jobs, netip.MustParsePrefix("52.21.153.129/32"))
	require.Contains(t, doc.Core, netip.MustParsePrefix("3.210.128.175/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("100.27.248.128/25"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/circleci.json")
	require.NoError(t, err)

	doc, err := circleci.ProcessData(data)
	require.NoError(t, err)
	require.NotEmpty(t, doc.MacOS)
	require.NotEmpty(t, doc.Jobs)
	require.NotEmpty(t, doc.Core)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := circleci.ProcessData([]byte("not json"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(circleci.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	c := circleci.New()
	gock.InterceptClient(c.Client.HTTPClient)

	_, err = c.Fetch()
	require.Error(t, err)
}
