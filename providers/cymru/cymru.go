package cymru

import (
	"bufio"
	"bytes"
	"encoding/json"
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
	ShortName = "cymru"
	FullName  = "Team Cymru Bogons"
	HostType  = "bogons"
	SourceURL = "https://www.team-cymru.com/bogon-reference"
	// IPv4URL and IPv6URL return every unallocated or reserved prefix, which
	// should never appear as a source address on the public internet.
	IPv4URL = "https://www.team-cymru.org/Services/Bogons/fullbogons-ipv4.txt"
	IPv6URL = "https://www.team-cymru.org/Services/Bogons/fullbogons-ipv6.txt"

	// lastUpdatedPrefix marks the header carrying the generation time, of the
	// form "# last updated 1787936101 (Fri Aug 28 16:55:01 2026 GMT)".
	lastUpdatedPrefix = "last updated"
)

type Cymru struct {
	Client  *retryablehttp.Client
	IPv4URL string
	IPv6URL string
	Timeout time.Duration
}

func New() Cymru {
	return Cymru{
		IPv4URL: IPv4URL,
		IPv6URL: IPv6URL,
		Client:  web.NewHTTPClientWithLogger(),
		Timeout: web.DefaultRequestTimeout,
	}
}

// RawDoc is the combined representation of the two upstream lists. It is the
// on-disk format so both address families are stored in a single,
// re-processable file.
type RawDoc struct {
	LastUpdated int64    `json:"lastUpdated,omitempty"`
	IPv4        []string `json:"ipv4"`
	IPv6        []string `json:"ipv6"`
}

type Doc struct {
	LastUpdated  time.Time      `json:"lastUpdated" yaml:"lastUpdated"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

// parseList splits a bogon list into its prefixes and the generation time from
// the header, which is a unix timestamp.
func parseList(data []byte) ([]string, int64) {
	var (
		prefixes    []string
		lastUpdated int64
	)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), bufio.MaxScanTokenSize)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			if _, value, found := strings.Cut(line, lastUpdatedPrefix); found {
				if ts, err := strconv.ParseInt(strings.Fields(value)[0], 10, 64); err == nil {
					lastUpdated = ts
				}
			}

			continue
		}

		prefixes = append(prefixes, line)
	}

	return prefixes, lastUpdated
}

func (c *Cymru) fetchList(url string) ([]string, int64, http.Header, int, error) {
	data, headers, status, err := web.Request(c.Client, url, http.MethodGet, nil, nil, c.Timeout)
	if err != nil {
		return nil, 0, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, 0, headers, status,
			fmt.Errorf("failed to download cymru bogons from %s. http status code: %d", url, status)
	}

	prefixes, lastUpdated := parseList(data)

	return prefixes, lastUpdated, headers, status, nil
}

func (c *Cymru) FetchData() ([]byte, http.Header, int, error) {
	if c.IPv4URL == "" {
		c.IPv4URL = IPv4URL
	}

	if c.IPv6URL == "" {
		c.IPv6URL = IPv6URL
	}

	v4, lastUpdated, headers, status, err := c.fetchList(c.IPv4URL)
	if err != nil {
		return nil, headers, status, err
	}

	v6, v6Updated, _, v6Status, err := c.fetchList(c.IPv6URL)
	if err != nil {
		return nil, headers, v6Status, err
	}

	// both lists are generated together, so either header will do.
	if lastUpdated == 0 {
		lastUpdated = v6Updated
	}

	combined, err := json.MarshalIndent(RawDoc{LastUpdated: lastUpdated, IPv4: v4, IPv6: v6}, "", " ")
	if err != nil {
		return nil, headers, status, err
	}

	return combined, headers, status, nil
}

func (c *Cymru) Fetch() (Doc, error) {
	data, _, _, err := c.FetchData()
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

	doc := Doc{
		IPv4Prefixes: castPrefixes(rawDoc.IPv4),
		IPv6Prefixes: castPrefixes(rawDoc.IPv6),
	}

	if rawDoc.LastUpdated != 0 {
		doc.LastUpdated = time.Unix(rawDoc.LastUpdated, 0).UTC()
	}

	return doc, nil
}

// castPrefixes parses the prefix strings, logging and skipping any that do not
// parse rather than discarding the whole list.
func castPrefixes(in []string) []netip.Prefix {
	var prefixes []netip.Prefix

	for _, entry := range in {
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			logrus.Warnf("failed to parse cymru prefix: %s", entry)

			continue
		}

		prefixes = append(prefixes, prefix)
	}

	return prefixes
}
