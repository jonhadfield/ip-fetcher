package dshield

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "dshield"
	FullName  = "DShield Recommended Block List"
	HostType  = "threat"
	SourceURL = "https://www.dshield.org/"
	// DownloadURL returns the top attacking networks over the last three days
	// as a tab delimited list, preceded by a commented header.
	DownloadURL = "https://feeds.dshield.org/block.txt"

	// fieldCount is the number of tab separated columns in a data row:
	// start, end, netmask, attacks, network name, country and abuse contact.
	fieldCount = 7

	// updatedPrefix marks the header comment carrying the generation time.
	updatedPrefix = "updated:"
)

// updatedFormat is the timestamp layout used by the header's updated comment.
// It carries no zone, and the feed is generated in UTC.
const updatedFormat = "2006-01-02T15:04:05.999999"

type DShield struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() DShield {
	return DShield{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

// Record is a single attacking network. Attacks is the number of targets that
// reported scans from the network, and Name, Country and Email are the
// registration details of the owning network, where DShield knows them.
type Record struct {
	Prefix  netip.Prefix `json:"prefix" yaml:"prefix"`
	Attacks int          `json:"attacks" yaml:"attacks"`
	Name    string       `json:"name,omitempty" yaml:"name,omitempty"`
	Country string       `json:"country,omitempty" yaml:"country,omitempty"`
	Email   string       `json:"email,omitempty" yaml:"email,omitempty"`
}

type Doc struct {
	Updated time.Time `json:"updated" yaml:"updated"`
	Records []Record  `json:"records" yaml:"records"`
}

func (d *DShield) FetchData() ([]byte, http.Header, int, error) {
	if d.DownloadURL == "" {
		d.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(d.Client, d.DownloadURL, http.MethodGet, nil, nil, d.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download dshield block list from %s. http status code: %d", d.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (d *DShield) Fetch() (Doc, error) {
	data, _, _, err := d.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	doc := Doc{}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), bufio.MaxScanTokenSize)

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")

		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			if updated, ok := parseUpdated(line); ok {
				doc.Updated = updated
			}

			continue
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		record, ok := parseRecord(line)
		if !ok {
			continue
		}

		doc.Records = append(doc.Records, record)
	}

	if err := scanner.Err(); err != nil {
		return Doc{}, err
	}

	return doc, nil
}

// parseUpdated extracts the generation time from the header comment of the
// form "#    updated: 2026-08-20T14:00:24.391841".
func parseUpdated(line string) (time.Time, bool) {
	_, value, found := strings.Cut(line, updatedPrefix)
	if !found {
		return time.Time{}, false
	}

	updated, err := time.Parse(updatedFormat, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, false
	}

	return updated.UTC(), true
}

// parseRecord converts a tab delimited data row into a Record. The start
// address and netmask give the prefix; the end address is redundant and so is
// only used to confirm the row is well formed.
func parseRecord(line string) (Record, bool) {
	fields := strings.Split(line, "\t")
	if len(fields) != fieldCount {
		logrus.Warnf("failed to parse dshield row, want %d fields: %s", fieldCount, line)

		return Record{}, false
	}

	addr, err := netip.ParseAddr(strings.TrimSpace(fields[0]))
	if err != nil {
		logrus.Warnf("failed to parse dshield network start: %s", fields[0])

		return Record{}, false
	}

	bits, err := strconv.Atoi(strings.TrimSpace(fields[2]))
	if err != nil {
		logrus.Warnf("failed to parse dshield netmask: %s", fields[2])

		return Record{}, false
	}

	prefix, err := addr.Prefix(bits)
	if err != nil {
		logrus.Warnf("failed to derive dshield prefix from %s/%d", addr, bits)

		return Record{}, false
	}

	// attacks is occasionally absent, in which case it is left at zero.
	attacks, err := strconv.Atoi(strings.TrimSpace(fields[3]))
	if err != nil {
		attacks = 0
	}

	return Record{
		Prefix:  prefix,
		Attacks: attacks,
		Name:    optional(fields[4]),
		Country: optional(fields[5]),
		Email:   optional(fields[6]),
	}, true
}

// optional normalises the placeholders DShield uses for unknown values.
func optional(in string) string {
	value := strings.TrimSpace(in)
	switch value {
	case "-", "None", ">>UNKNOWN<<":
		return ""
	default:
		return value
	}
}
