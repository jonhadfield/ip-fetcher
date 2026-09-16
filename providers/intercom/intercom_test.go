package intercom_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/intercom"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockIntercom(t *testing.T) *intercom.Intercom {
	t.Helper()

	for downloadURL, file := range map[string]string{
		intercom.USURL: "testdata/us.json",
		intercom.EUURL: "testdata/eu.json",
		intercom.AUURL: "testdata/au.json",
	} {
		u, err := url.Parse(downloadURL)
		require.NoError(t, err)

		gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
			Get(u.Path).
			Reply(http.StatusOK).
			File(file)
	}

	i := intercom.New()
	gock.InterceptClient(i.Client.HTTPClient)

	return &i
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	i := mockIntercom(t)

	doc, err := i.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("34.197.76.213/32"))
	require.NotEmpty(t, doc.IPRanges)
}

func TestFetchDataMergesRegions(t *testing.T) {
	defer gock.Off()

	i := mockIntercom(t)

	data, _, status, err := i.FetchData()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Contains(t, string(data), "INTERCOM-OUTBOUND")
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(intercom.USURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	i := intercom.New()
	gock.InterceptClient(i.Client.HTTPClient)

	_, err = i.Fetch()
	require.Error(t, err)
}
