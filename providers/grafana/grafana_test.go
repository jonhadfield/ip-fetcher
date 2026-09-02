package grafana_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/grafana"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockAllowlist(t *testing.T) *grafana.Grafana {
	t.Helper()

	u, err := url.Parse(grafana.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/synthetics.json")

	g := grafana.New()
	gock.InterceptClient(g.Client.HTTPClient)

	return &g
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	g := mockAllowlist(t)

	doc, err := g.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.Locations, 4)
	require.Len(t, doc.IPv4Prefixes, 10)
	require.Len(t, doc.IPv6Prefixes, 10)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("40.176.0.202/32"))
}

// map iteration order is random, so the locations must come back sorted.
func TestProcessDataSortsLocations(t *testing.T) {
	data, err := os.ReadFile("testdata/synthetics.json")
	require.NoError(t, err)

	doc, err := grafana.ProcessData(data)
	require.NoError(t, err)

	names := make([]string, 0, len(doc.Locations))
	for _, location := range doc.Locations {
		names = append(names, location.Name)
	}

	require.Equal(t, []string{"calgary", "capetown", "dublin", "frankfurt"}, names)
	require.NotEmpty(t, doc.Locations[0].IPv4)
	require.NotEmpty(t, doc.Locations[0].IPv6)
}

// the combined lists are the union of the locations', so an absent all member
// falls back to those.
func TestProcessDataFallsBackToLocations(t *testing.T) {
	raw := `{"locations":{"dublin":{"ipv4":["1.2.3.4/32"],"ipv6":["2001:db8::/56"]}}}`

	doc, err := grafana.ProcessData([]byte(raw))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("1.2.3.4/32")}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2001:db8::/56")}, doc.IPv6Prefixes)
}

// an unparseable prefix is skipped rather than failing the whole document.
func TestProcessDataSkipsInvalidPrefix(t *testing.T) {
	raw := `{"all":{"ipv4":["1.2.3.4/32","not-a-prefix"],"ipv6":[]}}`

	doc, err := grafana.ProcessData([]byte(raw))
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 1)
	require.Empty(t, doc.IPv6Prefixes)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := grafana.ProcessData([]byte("not json"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(grafana.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	g := grafana.New()
	gock.InterceptClient(g.Client.HTTPClient)

	_, err = g.Fetch()
	require.Error(t, err)
}
