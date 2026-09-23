package quiccloud_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/quiccloud"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(quiccloud.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/quiccloud.html")

	q := quiccloud.New()
	gock.InterceptClient(q.Client.HTTPClient)

	doc, err := q.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("102.221.36.98/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("103.167.151.84/32"))
}

func TestFindAddresses(t *testing.T) {
	addresses, err := quiccloud.FindAddresses([]byte("1.2.3.4<br />5.6.7.8/24<br />"))
	require.NoError(t, err)
	require.Equal(t, []string{"1.2.3.4", "5.6.7.8/24"}, addresses)

	doc, err := quiccloud.ProcessData([]byte(strings.Join(addresses, "\n") + "\n"))
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 2)
}

func TestFindAddressesEmpty(t *testing.T) {
	_, err := quiccloud.FindAddresses([]byte("<br />"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(quiccloud.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	q := quiccloud.New()
	gock.InterceptClient(q.Client.HTTPClient)

	_, err = q.Fetch()
	require.Error(t, err)
}
