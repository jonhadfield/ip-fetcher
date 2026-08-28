package cymru_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"
	"time"

	"github.com/jonhadfield/ip-fetcher/providers/cymru"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockLists(t *testing.T) {
	t.Helper()

	for _, l := range []struct{ raw, file string }{
		{cymru.IPv4URL, "testdata/fullbogons-ipv4.txt"},
		{cymru.IPv6URL, "testdata/fullbogons-ipv6.txt"},
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

	c := cymru.New()
	gock.InterceptClient(c.Client.HTTPClient)

	doc, err := c.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 25)
	require.Len(t, doc.IPv6Prefixes, 20)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("0.0.0.0/8"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("::/10"))
}

// the header carries a unix timestamp for when both lists were generated.
func TestFetchLastUpdated(t *testing.T) {
	defer gock.Off()

	mockLists(t)

	c := cymru.New()
	gock.InterceptClient(c.Client.HTTPClient)

	doc, err := c.Fetch()
	require.NoError(t, err)
	require.False(t, doc.LastUpdated.IsZero())
	require.Equal(t, time.UTC, doc.LastUpdated.Location())
}

func TestProcessData(t *testing.T) {
	raw := `{"lastUpdated":1787936101,"ipv4":["0.0.0.0/8","10.0.0.0/8"],"ipv6":["::/10"]}`

	doc, err := cymru.ProcessData([]byte(raw))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),
		netip.MustParsePrefix("10.0.0.0/8"),
	}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("::/10")}, doc.IPv6Prefixes)
	require.Equal(t, time.Unix(1787936101, 0).UTC(), doc.LastUpdated)
}

// an unparseable prefix is skipped rather than failing the whole list.
func TestProcessDataSkipsInvalidPrefix(t *testing.T) {
	doc, err := cymru.ProcessData([]byte(`{"ipv4":["0.0.0.0/8","nonsense"],"ipv6":[]}`))
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 1)
	require.True(t, doc.LastUpdated.IsZero())
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := cymru.ProcessData([]byte("not json"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(cymru.IPv4URL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	c := cymru.New()
	gock.InterceptClient(c.Client.HTTPClient)

	_, err = c.Fetch()
	require.Error(t, err)
}
