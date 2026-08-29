package newrelic_test

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/newrelic"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(newrelic.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ip-ranges.json")

	n := newrelic.New()
	gock.InterceptClient(n.Client.HTTPClient)

	doc, err := n.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.Locations, 5)
	require.Len(t, doc.IPv4Prefixes, 18)
	require.Empty(t, doc.IPv6Prefixes)
}

// the upstream document is a map, whose iteration order is random, so the
// locations must come out sorted for a stable document.
func TestProcessDataSortsLocations(t *testing.T) {
	data, err := os.ReadFile("testdata/ip-ranges.json")
	require.NoError(t, err)

	doc, err := newrelic.ProcessData(data)
	require.NoError(t, err)

	names := make([]string, 0, len(doc.Locations))
	for _, l := range doc.Locations {
		names = append(names, l.Name)
	}

	require.True(t, sort.StringsAreSorted(names), "locations should be sorted: %v", names)
}

// every prefix is attributed to the location that publishes it.
func TestProcessDataKeepsLocationGrouping(t *testing.T) {
	raw := `{"Berlin, DE":["1.2.3.0/24"],"Austin, TX, USA":["5.6.7.0/24","8.9.10.0/24"]}`

	doc, err := newrelic.ProcessData([]byte(raw))
	require.NoError(t, err)
	require.Len(t, doc.Locations, 2)
	require.Equal(t, "Austin, TX, USA", doc.Locations[0].Name)
	require.Len(t, doc.Locations[0].Prefixes, 2)
	require.Equal(t, "Berlin, DE", doc.Locations[1].Name)
	require.Len(t, doc.Locations[1].Prefixes, 1)
	require.Len(t, doc.IPv4Prefixes, 3)
}

// an unparseable entry is skipped rather than failing the whole document.
func TestProcessDataSkipsInvalidPrefix(t *testing.T) {
	doc, err := newrelic.ProcessData([]byte(`{"Nowhere":["1.2.3.0/24","nonsense"]}`))
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 1)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := newrelic.ProcessData([]byte("not json"))
	require.Error(t, err)
}
