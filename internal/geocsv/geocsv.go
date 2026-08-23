// Package geocsv parses the geolocation CSV published by providers that list a
// prefix alongside its country, region, city and postal code. Linode's geoip
// feed and iCloud Private Relay's egress ranges share this format.
package geocsv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/jszwec/csvutil"
)

const ipv6SeparatorThreshold = 2

// prefixPattern matches the address at the start of a line, before any comma.
var prefixPattern = regexp.MustCompile(`^[0-9a-fA-F]([^\s]+)`)

func IsIPv4(address string) bool {
	return strings.Count(address, ":") < ipv6SeparatorThreshold
}

func IsIPv6(address string) bool {
	return strings.Count(address, ":") >= ipv6SeparatorThreshold
}

// ExtractNet returns the CIDR at the start of in, adding a host mask to a bare
// address. It returns an empty string if there is no parseable network.
func ExtractNet(in string) string {
	s := prefixPattern.FindString(in)

	if !strings.Contains(s, "/") {
		switch {
		case IsIPv4(s):
			s += "/32"
		case IsIPv6(s):
			s += "/128"
		default:
			slog.Debug("failed to parse file line", "line", s)

			return ""
		}
	}

	if _, _, err := net.ParseCIDR(s); err != nil {
		slog.Debug("failed to parse file line", "line", s, "error", err)

		return ""
	}

	return s
}

// Record holds prefix, alpha2code, region, city and postal_code.
type Record struct {
	Prefix     netip.Prefix
	PrefixText string `csv:"ip_prefix,omitempty"`
	Alpha2Code string `csv:"alpha2code,omitempty"`
	Region     string `csv:"region,omitempty"`
	City       string `csv:"city,omitempty"`
	PostalCode string `csv:"postal_code,omitempty"`
}

// Entry is the CSV shape used to derive the header.
type Entry struct {
	Prefix     string `csv:"ip_prefix,omitempty"`
	Alpha2Code string `csv:"alpha2code,omitempty"`
	Region     string `csv:"region,omitempty"`
	City       string `csv:"city,omitempty"`
	PostalCode string `csv:"postal_code,omitempty"`
}

// CSVEntry mirrors Entry. Both providers exported it, so it is kept distinct
// to preserve their public API.
type CSVEntry struct {
	Prefix     string `csv:"ip_prefix,omitempty"`
	Alpha2Code string `csv:"alpha2code,omitempty"`
	Region     string `csv:"region,omitempty"`
	City       string `csv:"city,omitempty"`
	PostalCode string `csv:"postal_code,omitempty"`
}

type Doc struct {
	LastModified time.Time `json:"lastModified" yaml:"lastModified"`
	ETag         string    `json:"etag" yaml:"etag"`
	Records      []Record  `json:"records" yaml:"records"`
}

// FetchData returns the raw CSV document.
func FetchData(client *retryablehttp.Client, downloadURL string, timeout time.Duration) ([]byte, http.Header, int, error) {
	data, headers, status, err := web.Request(client, downloadURL, http.MethodGet, nil, nil, timeout)
	if status >= http.StatusBadRequest {
		return nil, nil, status, fmt.Errorf("failed to download prefixes. http status code: %d", status)
	}

	return data, headers, status, err
}

// Fetch returns the parsed document, carrying the ETag and Last-Modified
// values the provider served it with.
func Fetch(client *retryablehttp.Client, downloadURL string, timeout time.Duration) (Doc, error) {
	data, headers, _, err := FetchData(client, downloadURL, timeout)
	if err != nil {
		return Doc{}, err
	}

	records, err := Parse(data)
	if err != nil {
		return Doc{}, err
	}

	doc := Doc{Records: records}

	if etags := headers.Values(web.ETagHeader); len(etags) != 0 {
		doc.ETag = etags[0]
	}

	if lastModifiedRaw := headers.Values(web.LastModifiedHeader); len(lastModifiedRaw) != 0 {
		lastModified, perr := time.Parse(time.RFC1123, lastModifiedRaw[0])
		if perr != nil {
			return Doc{}, perr
		}

		doc.LastModified = lastModified
	}

	return doc, nil
}

func Parse(data []byte) ([]Record, error) {
	var records []Record

	csvReader := csv.NewReader(bytes.NewReader(data))
	csvReader.Comment = '#'
	csvReader.TrimLeadingSpace = true

	header, err := csvutil.Header(Entry{}, "csv")
	if err != nil {
		return records, err
	}

	dec, err := csvutil.NewDecoder(csvReader, header...)
	if err != nil {
		return records, err
	}

	for {
		var record Record

		err = dec.Decode(&record)

		switch {
		case errors.Is(err, io.EOF):
			return records, nil
		case err != nil:
			return records, err
		}

		if record.PrefixText == "" {
			continue
		}

		prefix, perr := netip.ParsePrefix(ExtractNet(record.PrefixText))
		if perr != nil {
			return records, perr
		}

		record.Prefix = prefix
		records = append(records, record)
	}
}
