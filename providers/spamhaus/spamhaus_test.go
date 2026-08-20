package spamhaus_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/spamhaus"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func mockLists(t *testing.T, s *spamhaus.Spamhaus) {
	t.Helper()

	v4, err := url.Parse(spamhaus.IPv4URL)
	require.NoError(t, err)
	gock.New(fmt.Sprintf("%s://%s", v4.Scheme, v4.Host)).
		Get(v4.Path).
		Reply(http.StatusOK).
		File("testdata/drop_v4.json")

	v6, err := url.Parse(spamhaus.IPv6URL)
	require.NoError(t, err)
	gock.New(fmt.Sprintf("%s://%s", v6.Scheme, v6.Host)).
		Get(v6.Path).
		Reply(http.StatusOK).
		File("testdata/drop_v6.json")

	gock.InterceptClient(s.Client.HTTPClient)
}

func TestFetch(t *testing.T) {
	defer gock.Off()

	s := spamhaus.New()
	mockLists(t, &s)

	doc, err := s.Fetch()
	require.NoError(t, err)

	require.Len(t, doc.IPv4Records, 20)
	require.Len(t, doc.IPv6Records, 10)

	require.Contains(t, doc.IPv4Records, spamhaus.Record{
		Prefix: netip.MustParsePrefix("1.10.16.0/20"),
		SBLID:  "SBL256894",
		RIR:    "apnic",
	})
	require.Contains(t, doc.IPv6Records, spamhaus.Record{
		Prefix: netip.MustParsePrefix("2001:678:254::/48"),
		SBLID:  "SBL697648",
		RIR:    "ripencc",
	})
}

// the trailing metadata line carries the generation time and must not be
// treated as a blocklist entry.
func TestFetchTimestampFromMetadata(t *testing.T) {
	defer gock.Off()

	s := spamhaus.New()
	mockLists(t, &s)

	doc, err := s.Fetch()
	require.NoError(t, err)
	require.Equal(t, int64(1786544042), doc.Timestamp.Unix())
}

// FetchData writes a single combined document that ProcessData can re-read.
func TestFetchDataRoundTrip(t *testing.T) {
	defer gock.Off()

	s := spamhaus.New()
	mockLists(t, &s)

	data, _, status, err := s.FetchData()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)

	var rawDoc spamhaus.RawDoc
	require.NoError(t, json.Unmarshal(data, &rawDoc))
	require.Len(t, rawDoc.IPv4, 21) // 20 entries plus the metadata record
	require.Len(t, rawDoc.IPv6, 11)

	doc, err := spamhaus.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Records, 20)
	require.Len(t, doc.IPv6Records, 10)
}

// unparseable prefixes are skipped rather than failing the whole list.
func TestProcessDataSkipsInvalidPrefix(t *testing.T) {
	raw, err := json.Marshal(spamhaus.RawDoc{
		IPv4: []spamhaus.RawRecord{
			{CIDR: "1.10.16.0/20", SBLID: "SBL256894", RIR: "apnic"},
			{CIDR: "not-a-prefix", SBLID: "SBL000000", RIR: "apnic"},
		},
	})
	require.NoError(t, err)

	doc, err := spamhaus.ProcessData(raw)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Records, 1)
	require.True(t, doc.Timestamp.IsZero())
}

// when only the IPv6 list carries metadata, its timestamp is used.
func TestProcessDataFallsBackToIPv6Timestamp(t *testing.T) {
	raw, err := json.Marshal(spamhaus.RawDoc{
		IPv4: []spamhaus.RawRecord{{CIDR: "1.10.16.0/20"}},
		IPv6: []spamhaus.RawRecord{{Type: "metadata", Timestamp: 1786542242}},
	})
	require.NoError(t, err)

	doc, err := spamhaus.ProcessData(raw)
	require.NoError(t, err)
	require.Equal(t, int64(1786542242), doc.Timestamp.Unix())
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := spamhaus.ProcessData([]byte(`not json`))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	s := spamhaus.New()

	v4, err := url.Parse(spamhaus.IPv4URL)
	require.NoError(t, err)
	gock.New(fmt.Sprintf("%s://%s", v4.Scheme, v4.Host)).
		Get(v4.Path).
		Reply(http.StatusNotFound)

	gock.InterceptClient(s.Client.HTTPClient)

	_, err = s.Fetch()
	require.Error(t, err)
}
