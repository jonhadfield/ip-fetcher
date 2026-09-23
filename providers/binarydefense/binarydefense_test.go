package binarydefense_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/binarydefense"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(binarydefense.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/banlist.txt")

	b := binarydefense.New()
	gock.InterceptClient(b.Client.HTTPClient)

	doc, err := b.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("1.20.168.127/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("203.0.113.10/32"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/banlist.txt")
	require.NoError(t, err)

	doc, err := binarydefense.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 4)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(binarydefense.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	b := binarydefense.New()
	gock.InterceptClient(b.Client.HTTPClient)

	_, err = b.Fetch()
	require.Error(t, err)
}
