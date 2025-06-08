package gcp_test

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/gcp"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(gcp.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		// SetHeader("Etag", "cd5e4f079775994d8e49f63ae9a84065").
		File("testdata/cloud.json")

	ac := gcp.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/cloud.json")
	require.NoError(t, err)

	doc, err := gcp.ProcessData(data)
	require.NoError(t, err)
	require.NotZero(t, doc.CreationTime)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.NotEmpty(t, doc.IPv6Prefixes)
}

func TestProcessDataBadPrefix(t *testing.T) {
	badJSON := []byte(`{"prefixes": [{"ipv4Prefix": "bad"}], "creationTime": "2022-07-15T10:03:23.044306"}`)
	_, err := gcp.ProcessData(badJSON)
	require.Error(t, err)
}
