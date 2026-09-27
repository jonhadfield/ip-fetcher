package ohdear_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/ohdear"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(ohdear.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ips.txt")

	p := ohdear.New()
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("45.32.146.84/32"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2001:19f0:6801:5f3:5400:1ff:fe38:3572/128"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/ips.txt")
	require.NoError(t, err)

	doc, err := ohdear.ProcessData(data)
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.NotEmpty(t, doc.IPv6Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(ohdear.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := ohdear.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
