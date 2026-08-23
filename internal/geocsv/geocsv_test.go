package geocsv_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"
	"time"

	"github.com/jonhadfield/ip-fetcher/internal/geocsv"
	"github.com/jonhadfield/ip-fetcher/internal/web"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

const testURL = "https://example.com/geofeed.csv"

func TestIsIPv4AndIsIPv6(t *testing.T) {
	cases := []struct {
		in string
		v4 bool
	}{
		{"1.2.3.4", true},
		{"1.2.3.0/24", true},
		{"2600:3c00::/32", false},
		{"2001:db8::1", false},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			require.Equal(t, tc.v4, geocsv.IsIPv4(tc.in))
			require.Equal(t, !tc.v4, geocsv.IsIPv6(tc.in))
		})
	}
}

// a bare address gains a host mask; a CIDR is returned as-is.
func TestExtractNet(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1.2.3.0/24", "1.2.3.0/24"},
		{"1.2.3.4", "1.2.3.4/32"},
		{"2600:3c00::/32", "2600:3c00::/32"},
		{"2001:db8::1", "2001:db8::1/128"},
		{"not-an-address", ""},
		{"", ""},
		{"999.1.1.1", ""},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			require.Equal(t, tc.want, geocsv.ExtractNet(tc.in))
		})
	}
}

// ExtractNet expects a single field, not a whole csv line: the pattern runs to
// the next whitespace, so trailing columns are swallowed and nothing parses.
// Parse only ever passes it the already split prefix column.
func TestExtractNetWholeLineYieldsNothing(t *testing.T) {
	require.Empty(t, geocsv.ExtractNet("1.2.3.0/24,US,US-TX,Richardson,"))
}

func TestParse(t *testing.T) {
	data := []byte("# comment\n" +
		"2600:3c00::/32,US,US-TX,Richardson,75081\n" +
		"1.2.3.0/24,GB,GB-EN,London,\n")

	records, err := geocsv.Parse(data)
	require.NoError(t, err)
	require.Len(t, records, 2)

	require.Equal(t, netip.MustParsePrefix("2600:3c00::/32"), records[0].Prefix)
	require.Equal(t, "US", records[0].Alpha2Code)
	require.Equal(t, "US-TX", records[0].Region)
	require.Equal(t, "Richardson", records[0].City)
	require.Equal(t, "75081", records[0].PostalCode)

	require.Equal(t, netip.MustParsePrefix("1.2.3.0/24"), records[1].Prefix)
	require.Equal(t, "London", records[1].City)
}

// rows without a prefix carry no address and are skipped.
func TestParseSkipsRowsWithoutPrefix(t *testing.T) {
	records, err := geocsv.Parse([]byte("1.2.3.0/24,GB,GB-EN,London,\n,US,US-TX,Richardson,\n"))
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, netip.MustParsePrefix("1.2.3.0/24"), records[0].Prefix)
}

func TestParseEmpty(t *testing.T) {
	records, err := geocsv.Parse(nil)
	require.NoError(t, err)
	require.Empty(t, records)
}

// an unparseable prefix is an error rather than a silently dropped row.
func TestParseInvalidPrefix(t *testing.T) {
	_, err := geocsv.Parse([]byte("not-an-address,GB,GB-EN,London,\n"))
	require.Error(t, err)
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(testURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		SetHeader(web.ETagHeader, `"abc123"`).
		SetHeader(web.LastModifiedHeader, "Mon, 02 Jan 2006 15:04:05 GMT").
		BodyString("1.2.3.0/24,GB,GB-EN,London,\n")

	client := web.NewHTTPClientWithLogger()
	gock.InterceptClient(client.HTTPClient)

	doc, err := geocsv.Fetch(client, testURL, web.DefaultRequestTimeout)
	require.NoError(t, err)
	require.Len(t, doc.Records, 1)
	require.Equal(t, `"abc123"`, doc.ETag)
	require.Equal(t,
		time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
		doc.LastModified.UTC())
}

// a response without the caching headers still yields records.
func TestFetchWithoutHeaders(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(testURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		BodyString("1.2.3.0/24,GB,GB-EN,London,\n")

	client := web.NewHTTPClientWithLogger()
	gock.InterceptClient(client.HTTPClient)

	doc, err := geocsv.Fetch(client, testURL, web.DefaultRequestTimeout)
	require.NoError(t, err)
	require.Len(t, doc.Records, 1)
	require.Empty(t, doc.ETag)
	require.True(t, doc.LastModified.IsZero())
}

// an unparseable Last-Modified is reported rather than ignored.
func TestFetchBadLastModified(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(testURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		SetHeader(web.LastModifiedHeader, "not a date").
		BodyString("1.2.3.0/24,GB,GB-EN,London,\n")

	client := web.NewHTTPClientWithLogger()
	gock.InterceptClient(client.HTTPClient)

	_, err = geocsv.Fetch(client, testURL, web.DefaultRequestTimeout)
	require.Error(t, err)
}

func TestFetchDataBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(testURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	client := web.NewHTTPClientWithLogger()
	gock.InterceptClient(client.HTTPClient)

	_, _, status, err := geocsv.FetchData(client, testURL, web.DefaultRequestTimeout)
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, status)
}
