package uptrends_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/uptrends"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockLists(t *testing.T) {
	t.Helper()

	for _, l := range []struct{ raw, file string }{
		{uptrends.IPv4URL, "testdata/ipv4.json"},
		{uptrends.IPv6URL, "testdata/ipv6.json"},
	} {
		u, err := url.Parse(l.raw)
		require.NoError(t, err)

		gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
			Get(u.Path).
			Reply(http.StatusOK).
			File(l.file)
	}
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	mockLists(t)

	u := uptrends.New()
	gock.InterceptClient(u.Client.HTTPClient)

	doc, err := u.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 15)
	require.Len(t, doc.IPv6Prefixes, 10)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("101.201.208.194/32"))
}

// both endpoints return bare addresses, which become host prefixes.
func TestFetchProducesHostPrefixes(t *testing.T) {
	defer gock.Off()

	mockLists(t)

	u := uptrends.New()
	gock.InterceptClient(u.Client.HTTPClient)

	doc, err := u.Fetch()
	require.NoError(t, err)

	for _, p := range doc.IPv4Prefixes {
		require.True(t, p.Addr().Is4())
		require.Equal(t, 32, p.Bits())
	}

	for _, p := range doc.IPv6Prefixes {
		require.True(t, p.Addr().Is6())
		require.Equal(t, 128, p.Bits())
	}
}

// FetchData combines the two responses into one re-processable document.
func TestFetchDataCombinesFamilies(t *testing.T) {
	defer gock.Off()

	mockLists(t)

	u := uptrends.New()
	gock.InterceptClient(u.Client.HTTPClient)

	data, _, _, err := u.FetchData()
	require.NoError(t, err)

	var raw uptrends.RawDoc
	require.NoError(t, json.Unmarshal(data, &raw))
	require.Len(t, raw.IPv4, 15)
	require.Len(t, raw.IPv6, 10)

	doc, err := uptrends.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 15)
	require.Len(t, doc.IPv6Prefixes, 10)
}

// an unparseable address is skipped rather than failing the whole list.
func TestProcessDataSkipsInvalidEntry(t *testing.T) {
	doc, err := uptrends.ProcessData([]byte(`{"ipv4":["1.2.3.4","not-an-ip"],"ipv6":[]}`))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("1.2.3.4/32")}, doc.IPv4Prefixes)
	require.Empty(t, doc.IPv6Prefixes)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := uptrends.ProcessData([]byte("not json"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(uptrends.IPv4URL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	ut := uptrends.New()
	gock.InterceptClient(ut.Client.HTTPClient)

	_, err = ut.Fetch()
	require.Error(t, err)
}
