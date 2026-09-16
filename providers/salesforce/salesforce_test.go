package salesforce_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/salesforce"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(salesforce.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/salesforce.json")

	s := salesforce.New()
	gock.InterceptClient(s.Client.HTTPClient)

	doc, err := s.Fetch()
	require.NoError(t, err)
	require.Equal(t, "1783342800", doc.SyncToken)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("145.224.193.0/24"))
	require.NotEmpty(t, doc.Regions)
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/salesforce.json")
	require.NoError(t, err)

	doc, err := salesforce.ProcessData(data)
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.Equal(t, "af-south-1", doc.Regions[0].Name)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(salesforce.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	s := salesforce.New()
	gock.InterceptClient(s.Client.HTTPClient)

	_, err = s.Fetch()
	require.Error(t, err)
}
