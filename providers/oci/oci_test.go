package oci_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/oci"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(oci.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/public_ip_ranges.json")

	ac := oci.New()
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.Regions)

	require.Len(t, doc.Regions, 3)
	require.Equal(t, "us-phoenix-1", doc.Regions[0].Region)
	require.Equal(t, "ap-mumbai-1", doc.Regions[2].Region)
	require.Equal(t, netip.MustParsePrefix("132.226.184.0/21"), doc.Regions[2].CIDRS[2].CIDR)
	require.Equal(t, "OCI", doc.Regions[2].CIDRS[2].Tags[0])
}
