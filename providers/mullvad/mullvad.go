package mullvad

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
	ShortName = "mullvad"
	FullName  = "Mullvad"
	HostType  = "anonymiser"
	SourceURL = "https://mullvad.net/en/servers"
	// DownloadURL returns Mullvad's full relay list, including each relay's
	// ingress addresses.
	DownloadURL = "https://api.mullvad.net/www/relays/all/"
)

type Mullvad struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Mullvad {
	return Mullvad{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Relay struct {
	Hostname    string `json:"hostname"`
	Active      bool   `json:"active"`
	IPv4AddrIn  string `json:"ipv4_addr_in"`
	IPv6AddrIn  string `json:"ipv6_addr_in"`
	CountryCode string `json:"country_code"`
	CityCode    string `json:"city_code"`
	Type        string `json:"type"`
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (m *Mullvad) FetchData() ([]byte, http.Header, int, error) {
	if m.DownloadURL == "" {
		m.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(m.Client, m.DownloadURL, http.MethodGet, nil, nil, m.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download mullvad relays from %s. http status code: %d", m.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (m *Mullvad) Fetch() (Doc, error) {
	data, _, _, err := m.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var relays []Relay
	if err := json.Unmarshal(data, &relays); err != nil {
		return Doc{}, err
	}

	seen4 := map[string]struct{}{}
	seen6 := map[string]struct{}{}

	var doc Doc

	for _, relay := range relays {
		if !relay.Active {
			continue
		}

		addAddr := func(entry string, seen map[string]struct{}, dest *[]netip.Prefix) {
			if entry == "" {
				return
			}

			prefix, ok := iplist.ToPrefix(entry)
			if !ok {
				return
			}

			if _, exists := seen[prefix.String()]; exists {
				return
			}

			seen[prefix.String()] = struct{}{}
			*dest = append(*dest, prefix)
		}

		addAddr(relay.IPv4AddrIn, seen4, &doc.IPv4Prefixes)
		addAddr(relay.IPv6AddrIn, seen6, &doc.IPv6Prefixes)
	}

	sort.Slice(doc.IPv4Prefixes, func(i, j int) bool {
		return doc.IPv4Prefixes[i].String() < doc.IPv4Prefixes[j].String()
	})
	sort.Slice(doc.IPv6Prefixes, func(i, j int) bool {
		return doc.IPv6Prefixes[i].String() < doc.IPv6Prefixes[j].String()
	})

	return doc, nil
}
