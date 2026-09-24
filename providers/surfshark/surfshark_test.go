package surfshark_test

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/surfshark"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(surfshark.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/clusters.json")

	p := surfshark.New()
	p.LookupIP = func(_ context.Context, host string) ([]netip.Addr, error) {
		switch host {
		case "al-tia.prod.surfshark.com":
			return []netip.Addr{netip.MustParseAddr("172.216.15.93")}, nil
		case "uk-lon.prod.surfshark.com":
			return []netip.Addr{
				netip.MustParseAddr("138.199.29.230"),
				netip.MustParseAddr("2a01:db8::1"),
			}, nil
		default:
			return nil, fmt.Errorf("unexpected host %s", host)
		}
	}
	gock.InterceptClient(p.Client.HTTPClient)

	doc, err := p.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("172.216.15.93/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("138.199.29.230/32"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2a01:db8::1/128"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/resolved.json")
	require.NoError(t, err)

	doc, err := surfshark.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 2)
	require.Len(t, doc.IPv6Prefixes, 1)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(surfshark.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	p := surfshark.New()
	gock.InterceptClient(p.Client.HTTPClient)

	_, err = p.Fetch()
	require.Error(t, err)
}
