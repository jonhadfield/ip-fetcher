package uptimerobot_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/uptimerobot"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(uptimerobot.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/uptimerobot.txt")

	ur := uptimerobot.New()
	gock.InterceptClient(ur.Client.HTTPClient)

	doc, err := ur.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 12)
	require.Len(t, doc.IPv6Prefixes, 4)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("3.12.251.153/32"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2607:ff68:107::33/128"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/uptimerobot.txt")
	require.NoError(t, err)

	doc, err := uptimerobot.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 12)
	require.Len(t, doc.IPv6Prefixes, 4)
}

// bare addresses become host prefixes, and CIDRs are accepted as-is.
func TestProcessDataMixedNotation(t *testing.T) {
	doc, err := uptimerobot.ProcessData([]byte("3.12.251.153\n\n10.0.0.0/8\n2607:ff68:107::33\n"))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("3.12.251.153/32"),
		netip.MustParsePrefix("10.0.0.0/8"),
	}, doc.IPv4Prefixes)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2607:ff68:107::33/128")}, doc.IPv6Prefixes)
}

// unparseable entries are skipped rather than failing the whole list.
func TestProcessDataSkipsInvalidEntry(t *testing.T) {
	doc, err := uptimerobot.ProcessData([]byte("3.12.251.153\nnot-an-ip\n"))
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 1)
	require.Empty(t, doc.IPv6Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(uptimerobot.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	ur := uptimerobot.New()
	gock.InterceptClient(ur.Client.HTTPClient)

	_, err = ur.Fetch()
	require.Error(t, err)
}
