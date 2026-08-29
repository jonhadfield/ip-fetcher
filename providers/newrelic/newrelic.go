package newrelic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"sort"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "newrelic"
	FullName  = "New Relic Synthetics"
	HostType  = "monitoring"
	SourceURL = "https://docs.newrelic.com/docs/synthetics/synthetic-monitoring/administration/synthetic-public-minion-ips/"
	// DownloadURL returns the ranges New Relic's public synthetics monitors run
	// from, as json keyed by location name.
	DownloadURL = "https://s3.amazonaws.com/nr-synthetics-assets/nat-ip-dnsname/production/ip-ranges.json"
)

type NewRelic struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() NewRelic {
	return NewRelic{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

// Location groups the prefixes a named monitoring location runs from.
type Location struct {
	Name     string         `json:"name" yaml:"name"`
	Prefixes []netip.Prefix `json:"prefixes" yaml:"prefixes"`
}

type Doc struct {
	Locations    []Location     `json:"locations" yaml:"locations"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (n *NewRelic) FetchData() ([]byte, http.Header, int, error) {
	if n.DownloadURL == "" {
		n.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(n.Client, n.DownloadURL, http.MethodGet, nil, nil, n.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download newrelic ranges from %s. http status code: %d", n.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (n *NewRelic) Fetch() (Doc, error) {
	data, _, _, err := n.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	// the document is keyed by location name, such as "Washington, DC, USA".
	var raw map[string][]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	// map iteration order is random, so sort for a stable document.
	names := make([]string, 0, len(raw))
	for name := range raw {
		names = append(names, name)
	}

	sort.Strings(names)

	doc := Doc{Locations: make([]Location, 0, len(names))}

	for _, name := range names {
		location := Location{Name: name}

		for _, entry := range raw[name] {
			prefix, err := netip.ParsePrefix(entry)
			if err != nil {
				logrus.Warnf("failed to parse newrelic prefix: %s", entry)

				continue
			}

			location.Prefixes = append(location.Prefixes, prefix)

			if prefix.Addr().Is4() {
				doc.IPv4Prefixes = append(doc.IPv4Prefixes, prefix)

				continue
			}

			doc.IPv6Prefixes = append(doc.IPv6Prefixes, prefix)
		}

		doc.Locations = append(doc.Locations, location)
	}

	return doc, nil
}
