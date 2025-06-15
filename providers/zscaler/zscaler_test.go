package zscaler_test

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/zscaler"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/doc.json")
	require.NoError(t, err)
	require.NotEmpty(t, data)

	doc, err := zscaler.ProcessData(data)
	require.NoError(t, err)
	require.Equal(t, "87.58.112.0/23", doc.ZscalerNet.ContinentEMEA.CityAmsterdamIII[1].Range)
}

func TestFetch(t *testing.T) {
	u, err := url.Parse(zscaler.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/doc.json")

	z := zscaler.New()
	gock.InterceptClient(z.Client.HTTPClient)

	doc, err := z.Fetch()
	require.NoError(t, err)
	require.Equal(t, "87.58.112.0/23", doc.ZscalerNet.ContinentEMEA.CityAmsterdamIII[1].Range)
}
