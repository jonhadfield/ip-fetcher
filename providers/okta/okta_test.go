package okta_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"sort"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/okta"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockOkta(t *testing.T) *okta.Okta {
	t.Helper()

	u, err := url.Parse(okta.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/ip_ranges.json")

	o := okta.New()
	gock.InterceptClient(o.Client.HTTPClient)

	return &o
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	o := mockOkta(t)

	doc, err := o.Fetch()
	require.NoError(t, err)
	require.Len(t, doc.Cells, 4)
}

// an org is hosted on one cell, so the cell name must survive as what
// identifies the ranges to allowlist.
func TestFetchKeepsCellNames(t *testing.T) {
	defer gock.Off()

	o := mockOkta(t)

	doc, err := o.Fetch()
	require.NoError(t, err)

	names := make([]string, 0, len(doc.Cells))
	for _, cell := range doc.Cells {
		names = append(names, cell.Name)
	}

	require.Equal(t, []string{"apac_cell_1", "emea_cell_1", "preview_cell_1", "us_cell_1"}, names)
}

// the document is a JSON object, so the order must be imposed rather than
// inherited, or the output changes between runs on identical input.
func TestProcessDataSortsCells(t *testing.T) {
	data, err := os.ReadFile("testdata/ip_ranges.json")
	require.NoError(t, err)

	doc, err := okta.ProcessData(data)
	require.NoError(t, err)

	names := make([]string, 0, len(doc.Cells))
	for _, cell := range doc.Cells {
		names = append(names, cell.Name)
	}

	require.True(t, sort.StringsAreSorted(names))
}

func TestProcessDataSplitsFamilies(t *testing.T) {
	data, err := os.ReadFile("testdata/ip_ranges.json")
	require.NoError(t, err)

	doc, err := okta.ProcessData(data)
	require.NoError(t, err)

	for _, cell := range doc.Cells {
		for _, prefix := range cell.IPv4Prefixes {
			require.True(t, prefix.Addr().Is4())
		}

		for _, prefix := range cell.IPv6Prefixes {
			require.True(t, prefix.Addr().Is6())
		}
	}

	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2600:1f14:1::/48")}, doc.Cells[0].IPv6Prefixes)
}

// an unparseable range is skipped rather than discarding its cell.
func TestProcessDataSkipsInvalidRange(t *testing.T) {
	doc, err := okta.ProcessData([]byte(`{"us_cell_1":{"ip_ranges":["1.2.3.0/24","not-a-prefix"]}}`))
	require.NoError(t, err)
	require.Len(t, doc.Cells, 1)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("1.2.3.0/24")}, doc.Cells[0].IPv4Prefixes)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := okta.ProcessData([]byte("not json"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(okta.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	o := okta.New()
	gock.InterceptClient(o.Client.HTTPClient)

	_, err = o.Fetch()
	require.Error(t, err)
}
