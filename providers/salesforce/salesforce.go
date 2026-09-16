package salesforce

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
	ShortName = "salesforce"
	FullName  = "Salesforce"
	HostType  = "saas"
	SourceURL = "https://help.salesforce.com/s/articleView?id=000384438&type=1"
	// DownloadURL is Salesforce's Hyperforce public IP ranges document.
	DownloadURL = "https://ip-ranges.salesforce.com/ip-ranges.json"
)

type Salesforce struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Salesforce {
	return Salesforce{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type RawDoc struct {
	SyncToken  string      `json:"syncToken"`
	CreateDate string      `json:"createDate"`
	Prefixes   []RawPrefix `json:"prefixes"`
}

// RawPrefix is one region's entry. Salesforce publishes ip_prefix as an array
// of CIDRs rather than a scalar, unlike the AWS schema it otherwise follows.
type RawPrefix struct {
	Region   string   `json:"region"`
	Provider string   `json:"provider"`
	IPPrefix []string `json:"ip_prefix"`
}

type Region struct {
	Name         string         `json:"name"          yaml:"name"`
	Provider     string         `json:"provider"      yaml:"provider"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

type Doc struct {
	SyncToken    string         `json:"syncToken"     yaml:"syncToken"`
	CreateDate   string         `json:"createDate"    yaml:"createDate"`
	Regions      []Region       `json:"regions"       yaml:"regions"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (s *Salesforce) FetchData() ([]byte, http.Header, int, error) {
	if s.DownloadURL == "" {
		s.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(s.Client, s.DownloadURL, http.MethodGet, nil, nil, s.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download salesforce ranges from %s. http status code: %d", s.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (s *Salesforce) Fetch() (Doc, error) {
	data, _, _, err := s.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var raw RawDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	doc := Doc{
		SyncToken:  raw.SyncToken,
		CreateDate: raw.CreateDate,
	}

	seen4 := map[string]struct{}{}
	seen6 := map[string]struct{}{}

	for _, entry := range raw.Prefixes {
		region := Region{Name: entry.Region, Provider: entry.Provider}

		for _, prefix := range iplist.CastPrefixes(ShortName+" "+entry.Region, entry.IPPrefix) {
			if prefix.Addr().Is4() {
				region.IPv4Prefixes = append(region.IPv4Prefixes, prefix)
				if _, ok := seen4[prefix.String()]; !ok {
					seen4[prefix.String()] = struct{}{}
					doc.IPv4Prefixes = append(doc.IPv4Prefixes, prefix)
				}

				continue
			}

			region.IPv6Prefixes = append(region.IPv6Prefixes, prefix)
			if _, ok := seen6[prefix.String()]; !ok {
				seen6[prefix.String()] = struct{}{}
				doc.IPv6Prefixes = append(doc.IPv6Prefixes, prefix)
			}
		}

		doc.Regions = append(doc.Regions, region)
	}

	sort.Slice(doc.IPv4Prefixes, func(i, j int) bool {
		return doc.IPv4Prefixes[i].String() < doc.IPv4Prefixes[j].String()
	})
	sort.Slice(doc.IPv6Prefixes, func(i, j int) bool {
		return doc.IPv6Prefixes[i].String() < doc.IPv6Prefixes[j].String()
	})

	return doc, nil
}
