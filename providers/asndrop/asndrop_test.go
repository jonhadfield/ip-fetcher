package asndrop_test

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/asndrop"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(asndrop.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/asndrop.json")

	p := asndrop.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.Records, 2)
	require.Equal(t, 245, doc.Records[0].ASN)
	require.Equal(t, "PRC-AS", doc.Records[0].ASName)
	require.False(t, doc.Timestamp.IsZero())
	require.Contains(t, string(doc.Lines()), "AS245")
	require.Contains(t, string(doc.Lines()), "AS2601")
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/asndrop.json")
	require.NoError(t, err)

	doc, err := asndrop.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.Records, 2)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(asndrop.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := asndrop.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
