package updown_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/updown"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockNodes(t *testing.T) *updown.Updown {
	t.Helper()

	u, err := url.Parse(updown.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/nodes.json")

	d := updown.New()
	gock.InterceptClient(d.Client.HTTPClient)

	return &d
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	d := mockNodes(t)

	doc, err := d.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.Nodes, 11)
	require.Len(t, doc.IPv4Prefixes, 11)
	require.Len(t, doc.IPv6Prefixes, 11)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("45.32.74.41/32"))
}

// the node's key becomes its name, and its location is carried through.
func TestProcessDataKeepsNodeDetail(t *testing.T) {
	data, err := os.ReadFile("testdata/nodes.json")
	require.NoError(t, err)

	doc, err := updown.ProcessData(data)
	require.NoError(t, err)

	var lan updown.Node

	for _, node := range doc.Nodes {
		if node.Name == "lan" {
			lan = node
		}
	}

	require.Equal(t, "lan", lan.Name)
	require.Equal(t, "45.32.74.41", lan.IP)
	require.Equal(t, "Los Angeles", lan.City)
	require.Equal(t, "us", lan.CountryCode)
	require.InDelta(t, 34.0729, lan.Latitude, 0.0001)
}

// map iteration order is random, so the nodes must come back sorted by name.
func TestProcessDataSortsNodes(t *testing.T) {
	data, err := os.ReadFile("testdata/nodes.json")
	require.NoError(t, err)

	doc, err := updown.ProcessData(data)
	require.NoError(t, err)

	for i := 1; i < len(doc.Nodes); i++ {
		require.Less(t, doc.Nodes[i-1].Name, doc.Nodes[i].Name)
	}
}

// a node without an IPv6 address must not become a bogus prefix, and an
// unparseable address is skipped rather than failing the whole document.
func TestProcessDataSkipsEmptyAndInvalid(t *testing.T) {
	raw := `{"a":{"ip":"1.2.3.4","ip6":""},"b":{"ip":"not-an-ip","ip6":"2001:db8::1"}}`

	doc, err := updown.ProcessData([]byte(raw))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("1.2.3.4/32")}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2001:db8::1/128")}, doc.IPv6Prefixes)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := updown.ProcessData([]byte("not json"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(updown.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	d := updown.New()
	gock.InterceptClient(d.Client.HTTPClient)

	_, err = d.Fetch()
	require.Error(t, err)
}
