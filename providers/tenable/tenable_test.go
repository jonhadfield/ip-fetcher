package tenable_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/tenable"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockRanges(t *testing.T) *tenable.Tenable {
	t.Helper()

	u, err := url.Parse(tenable.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/data.json")

	ten := tenable.New()
	gock.InterceptClient(ten.Client.HTTPClient)

	return &ten
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	ten := mockRanges(t)

	doc, err := ten.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.Prefixes, 4)
	require.Len(t, doc.IPv6Prefixes, 3)
	require.NotEmpty(t, doc.SyncToken)
	require.NotEmpty(t, doc.CreateDate)
}

// the sensor group names the scanners the prefix belongs to, and is what the
// region alone does not say.
func TestFetchKeepsSensorDetail(t *testing.T) {
	defer gock.Off()

	ten := mockRanges(t)

	doc, err := ten.Fetch()
	require.NoError(t, err)
	require.Equal(t, netip.MustParsePrefix("13.115.104.128/25"), doc.Prefixes[0].IPPrefix)
	require.Equal(t, "ap-northeast-1", doc.Prefixes[0].Region)
	require.Equal(t, "tenable-scanners", doc.Prefixes[0].Service)
	require.Contains(t, doc.Prefixes[0].SensorGroup, "AP Tokyo Cloud Scanners")
}

// only FedRAMP customers are scanned from the FedRAMP scanners, so they stay in
// members of their own.
func TestProcessDataSeparatesFedRAMP(t *testing.T) {
	data, err := os.ReadFile("testdata/data.json")
	require.NoError(t, err)

	doc, err := tenable.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.FedRAMPPrefixes, 2)
	require.Len(t, doc.FedRAMPIPv6Prefixes, 2)
	require.Equal(t, netip.MustParsePrefix("52.61.37.84/32"), doc.FedRAMPPrefixes[0].IPPrefix)

	for _, prefix := range doc.Prefixes {
		require.NotContains(t, doc.FedRAMPPrefixes, prefix)
	}
}

// each family lands in its own member, as its own prefixes.
func TestProcessDataSplitsFamilies(t *testing.T) {
	data, err := os.ReadFile("testdata/data.json")
	require.NoError(t, err)

	doc, err := tenable.ProcessData(data)
	require.NoError(t, err)

	for _, prefix := range doc.Prefixes {
		require.True(t, prefix.IPPrefix.Addr().Is4())
	}

	for _, prefix := range doc.IPv6Prefixes {
		require.True(t, prefix.IPPrefix.Addr().Is6())
	}
}

// an unparseable prefix is skipped rather than failing the whole document.
func TestProcessDataSkipsInvalidPrefix(t *testing.T) {
	raw := `{"prefixes":[{"ip_prefix":"1.2.3.0/24"},{"ip_prefix":"not-a-prefix"}],"ipv6_prefixes":[]}`

	doc, err := tenable.ProcessData([]byte(raw))
	require.NoError(t, err)
	require.Len(t, doc.Prefixes, 1)
	require.Empty(t, doc.IPv6Prefixes)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := tenable.ProcessData([]byte("not json"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(tenable.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	ten := tenable.New()
	gock.InterceptClient(ten.Client.HTTPClient)

	_, err = ten.Fetch()
	require.Error(t, err)
}
