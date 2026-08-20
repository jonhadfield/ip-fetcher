package spamhaus

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "spamhaus"
	FullName  = "Spamhaus DROP"
	HostType  = "threat"
	SourceURL = "https://www.spamhaus.org/blocklists/do-not-route-or-peer/"
	// IPv4URL returns the DROP list as newline delimited JSON objects.
	IPv4URL = "https://www.spamhaus.org/drop/drop_v4.json"
	// IPv6URL returns the IPv6 DROP list as newline delimited JSON objects.
	IPv6URL = "https://www.spamhaus.org/drop/drop_v6.json"

	metadataRecordType = "metadata"
)

type Spamhaus struct {
	Client  *retryablehttp.Client
	IPv4URL string
	IPv6URL string
	Timeout time.Duration
}

func New() Spamhaus {
	return Spamhaus{
		IPv4URL: IPv4URL,
		IPv6URL: IPv6URL,
		Client:  web.NewHTTPClientWithLogger(),
		Timeout: web.DefaultRequestTimeout,
	}
}

// RawRecord is a single line of the upstream newline delimited JSON. Blocklist
// entries carry cidr, sblid and rir, whilst the trailing metadata line carries
// the generation timestamp and record count.
type RawRecord struct {
	Type      string `json:"type,omitempty"`
	CIDR      string `json:"cidr,omitempty"`
	SBLID     string `json:"sblid,omitempty"`
	RIR       string `json:"rir,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Records   int    `json:"records,omitempty"`
}

// RawDoc is the combined representation of the two upstream lists. It is the
// on-disk format so that both address families are stored in a single,
// re-processable file.
type RawDoc struct {
	IPv4 []RawRecord `json:"ipv4"`
	IPv6 []RawRecord `json:"ipv6"`
}

type Record struct {
	Prefix netip.Prefix `json:"prefix" yaml:"prefix"`
	SBLID  string       `json:"sblid" yaml:"sblid"`
	RIR    string       `json:"rir" yaml:"rir"`
}

type Doc struct {
	Timestamp   time.Time `json:"timestamp" yaml:"timestamp"`
	IPv4Records []Record  `json:"ipv4_records" yaml:"ipv4_records"`
	IPv6Records []Record  `json:"ipv6_records" yaml:"ipv6_records"`
}

func (s *Spamhaus) fetchList(url string) ([]RawRecord, http.Header, int, error) {
	data, headers, status, err := web.Request(s.Client, url, http.MethodGet, nil, nil, s.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download spamhaus drop list from %s. http status code: %d", url, status)
	}

	records, err := parseLines(data)
	if err != nil {
		return nil, headers, status, err
	}

	return records, headers, status, nil
}

// parseLines decodes the newline delimited JSON body, ignoring blank lines.
func parseLines(data []byte) ([]RawRecord, error) {
	var records []RawRecord

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), bufio.MaxScanTokenSize)

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var record RawRecord
		if err := json.Unmarshal(line, &record); err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (s *Spamhaus) FetchData() ([]byte, http.Header, int, error) {
	if s.IPv4URL == "" {
		s.IPv4URL = IPv4URL
	}

	if s.IPv6URL == "" {
		s.IPv6URL = IPv6URL
	}

	v4, headers, status, err := s.fetchList(s.IPv4URL)
	if err != nil {
		return nil, headers, status, err
	}

	v6, _, v6status, err := s.fetchList(s.IPv6URL)
	if err != nil {
		return nil, headers, v6status, err
	}

	combined, err := json.MarshalIndent(RawDoc{IPv4: v4, IPv6: v6}, "", " ")
	if err != nil {
		return nil, headers, status, err
	}

	return combined, headers, status, nil
}

func (s *Spamhaus) Fetch() (Doc, error) {
	data, _, _, err := s.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var rawDoc RawDoc
	if err := json.Unmarshal(data, &rawDoc); err != nil {
		return Doc{}, err
	}

	doc := Doc{}

	var v4Timestamp, v6Timestamp int64

	doc.IPv4Records, v4Timestamp = castRecords(rawDoc.IPv4)
	doc.IPv6Records, v6Timestamp = castRecords(rawDoc.IPv6)

	// prefer the IPv4 list's generation time, as it is the larger of the two.
	timestamp := v4Timestamp
	if timestamp == 0 {
		timestamp = v6Timestamp
	}

	if timestamp != 0 {
		doc.Timestamp = time.Unix(timestamp, 0).UTC()
	}

	return doc, nil
}

// castRecords converts blocklist entries to Records and returns the generation
// timestamp carried by the list's trailing metadata record, if present.
func castRecords(raw []RawRecord) ([]Record, int64) {
	var (
		records   []Record
		timestamp int64
	)

	for _, r := range raw {
		if r.Type == metadataRecordType {
			timestamp = r.Timestamp

			continue
		}

		if r.CIDR == "" {
			continue
		}

		prefix, err := netip.ParsePrefix(r.CIDR)
		if err != nil {
			logrus.Warnf("failed to parse spamhaus prefix: %s", r.CIDR)

			continue
		}

		records = append(records, Record{
			Prefix: prefix,
			SBLID:  r.SBLID,
			RIR:    r.RIR,
		})
	}

	return records, timestamp
}
