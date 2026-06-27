package datadog

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
	ShortName   = "datadog"
	FullName    = "Datadog"
	HostType    = "saas"
	SourceURL   = "https://docs.datadoghq.com/api/latest/ip-ranges/"
	DownloadURL = "https://ip-ranges.datadoghq.com/"
)

type Datadog struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Datadog {
	return Datadog{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

// Category holds the prefixes for a single Datadog service (agents, api, ...).
type Category struct {
	IPv4Prefixes []netip.Prefix `json:"prefixes_ipv4" yaml:"prefixes_ipv4"`
	IPv6Prefixes []netip.Prefix `json:"prefixes_ipv6" yaml:"prefixes_ipv6"`
}

type Doc struct {
	Version      int                 `json:"version"       yaml:"version"`
	Modified     string              `json:"modified"      yaml:"modified"`
	Categories   map[string]Category `json:"categories"    yaml:"categories"`
	IPv4Prefixes []netip.Prefix      `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix      `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (d *Datadog) FetchData() ([]byte, http.Header, int, error) {
	if d.DownloadURL == "" {
		d.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(d.Client, d.DownloadURL, http.MethodGet, nil, nil, d.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status, fmt.Errorf("failed to download datadog prefixes. http status code: %d", status)
	}

	return data, headers, status, nil
}

func (d *Datadog) Fetch() (Doc, error) {
	data, _, _, err := d.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

// rawCategory mirrors the per-service object in the upstream document.
type rawCategory struct {
	PrefixesIPv4 []string `json:"prefixes_ipv4"`
	PrefixesIPv6 []string `json:"prefixes_ipv6"`
}

func ProcessData(data []byte) (Doc, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	doc := Doc{Categories: map[string]Category{}}

	if v, ok := raw["version"]; ok {
		_ = json.Unmarshal(v, &doc.Version)
	}

	if m, ok := raw["modified"]; ok {
		_ = json.Unmarshal(m, &doc.Modified)
	}

	seen4 := map[string]bool{}
	seen6 := map[string]bool{}

	for key, rawVal := range raw {
		if key == "version" || key == "modified" {
			continue
		}

		var rc rawCategory
		if err := json.Unmarshal(rawVal, &rc); err != nil {
			// not a service category (e.g. an unexpected scalar field) - skip it
			continue
		}

		doc.Categories[key] = Category{
			IPv4Prefixes: parsePrefixes(rc.PrefixesIPv4, seen4, &doc.IPv4Prefixes),
			IPv6Prefixes: parsePrefixes(rc.PrefixesIPv6, seen6, &doc.IPv6Prefixes),
		}
	}

	// Sort the flattened unions for deterministic output (map iteration is random).
	sortPrefixes(doc.IPv4Prefixes)
	sortPrefixes(doc.IPv6Prefixes)

	return doc, nil
}

// parsePrefixes parses the prefix strings for one category, appending any newly
// seen prefix to the document-wide union pointed at by total.
func parsePrefixes(entries []string, seen map[string]bool, total *[]netip.Prefix) []netip.Prefix {
	var prefixes []netip.Prefix

	for _, s := range entries {
		p, err := netip.ParsePrefix(s)
		if err != nil {
			logrus.Warnf("failed to parse datadog prefix: %s", s)

			continue
		}

		prefixes = append(prefixes, p)

		if !seen[p.String()] {
			seen[p.String()] = true
			*total = append(*total, p)
		}
	}

	return prefixes
}

func sortPrefixes(prefixes []netip.Prefix) {
	sort.Slice(prefixes, func(i, j int) bool {
		return prefixes[i].String() < prefixes[j].String()
	})
}
