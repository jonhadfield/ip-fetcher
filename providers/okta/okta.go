// Package okta retrieves the addresses Okta's identity service connects from,
// which are the addresses an org allowlists for inbound Okta traffic.
package okta

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"sort"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "okta"
	FullName  = "Okta"
	HostType  = "saas"
	SourceURL = "https://help.okta.com/en-us/content/topics/security/ip-address-allow-listing.htm"
	// DownloadURL returns Okta's addresses grouped by the cell serving them.
	// An org is hosted on one cell, so allowlisting usually needs the ranges of
	// that cell alone rather than the whole document.
	DownloadURL = "https://s3.amazonaws.com/okta-ip-ranges/ip_ranges.json"
)

type Okta struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Okta {
	return Okta{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

// RawDoc maps a cell name, such as us_cell_1, to that cell's ranges.
type RawDoc map[string]RawCell

type RawCell struct {
	IPRanges []string `json:"ip_ranges"`
}

// Cell names the ranges it holds, as the name is what identifies the cell an
// org is hosted on.
type Cell struct {
	Name         string         `json:"name"          yaml:"name"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

type Doc struct {
	Cells []Cell `json:"cells" yaml:"cells"`
}

func (o *Okta) FetchData() ([]byte, http.Header, int, error) {
	if o.DownloadURL == "" {
		o.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(o.Client, o.DownloadURL, http.MethodGet, nil, nil, o.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download okta ranges from %s. http status code: %d", o.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (o *Okta) Fetch() (Doc, error) {
	data, _, _, err := o.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

// ProcessData returns the cells sorted by name, as the document is a JSON
// object and so carries no order of its own.
func ProcessData(data []byte) (Doc, error) {
	var raw RawDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	names := make([]string, 0, len(raw))
	for name := range raw {
		names = append(names, name)
	}

	sort.Strings(names)

	cells := make([]Cell, 0, len(names))

	for _, name := range names {
		ipv4, ipv6 := splitFamilies(name, raw[name].IPRanges)

		cells = append(cells, Cell{Name: name, IPv4Prefixes: ipv4, IPv6Prefixes: ipv6})
	}

	if len(cells) == 0 {
		return Doc{}, nil
	}

	return Doc{Cells: cells}, nil
}

// splitFamilies sorts a cell's ranges into families, logging and skipping any
// entry that cannot be parsed so one bad range does not discard the cell.
func splitFamilies(name string, ranges []string) ([]netip.Prefix, []netip.Prefix) {
	var ipv4, ipv6 []netip.Prefix

	for _, prefix := range iplist.CastPrefixes(ShortName+" cell "+name, ranges) {
		if prefix.Addr().Is4() {
			ipv4 = append(ipv4, prefix)

			continue
		}

		ipv6 = append(ipv6, prefix)
	}

	return ipv4, ipv6
}
