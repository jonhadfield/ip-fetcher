package grafana

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
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "grafana"
	FullName  = "Grafana Synthetic Monitoring"
	HostType  = "monitoring"
	SourceURL = "https://grafana.com/docs/grafana-cloud/testing/synthetic-monitoring/create-checks/public-probes/"
	// DownloadURL returns the public probe ranges as a json document holding the
	// combined list and the same prefixes broken down by probe location.
	DownloadURL = "https://allowlists.grafana.com/synthetics"
)

type Grafana struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Grafana {
	return Grafana{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

// rawDoc is the upstream document: the combined ranges, then the same ranges
// keyed by probe location.
type rawDoc struct {
	All       ranges            `json:"all"`
	Locations map[string]ranges `json:"locations"`
}

type ranges struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

// Location is a single probe location, carrying its prefixes as published.
type Location struct {
	Name string   `json:"name" yaml:"name"`
	IPv4 []string `json:"ipv4" yaml:"ipv4"`
	IPv6 []string `json:"ipv6" yaml:"ipv6"`
}

type Doc struct {
	Locations    []Location     `json:"locations" yaml:"locations"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (g *Grafana) FetchData() ([]byte, http.Header, int, error) {
	if g.DownloadURL == "" {
		g.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(g.Client, g.DownloadURL, http.MethodGet, nil, nil, g.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download grafana allowlist from %s. http status code: %d", g.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (g *Grafana) Fetch() (Doc, error) {
	data, _, _, err := g.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var raw rawDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	doc := Doc{Locations: make([]Location, 0, len(raw.Locations))}

	for name, r := range raw.Locations {
		doc.Locations = append(doc.Locations, Location{Name: name, IPv4: r.IPv4, IPv6: r.IPv6})
	}

	// map iteration order is random, so sort to keep the document stable.
	sort.Slice(doc.Locations, func(i, j int) bool { return doc.Locations[i].Name < doc.Locations[j].Name })

	combined := raw.All
	// the combined lists are the union of the locations', so fall back to those
	// should the all member ever be absent.
	if len(combined.IPv4) == 0 && len(combined.IPv6) == 0 {
		for _, location := range doc.Locations {
			combined.IPv4 = append(combined.IPv4, location.IPv4...)
			combined.IPv6 = append(combined.IPv6, location.IPv6...)
		}
	}

	doc.IPv4Prefixes = castPrefixes(combined.IPv4)
	doc.IPv6Prefixes = castPrefixes(combined.IPv6)

	return doc, nil
}

func castPrefixes(in []string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(in))

	for _, entry := range in {
		prefix, ok := iplist.ToPrefix(entry)
		if !ok {
			logrus.Warnf("failed to parse grafana prefix: %s", entry)

			continue
		}

		prefixes = append(prefixes, prefix)
	}

	if len(prefixes) == 0 {
		return nil
	}

	return prefixes
}
