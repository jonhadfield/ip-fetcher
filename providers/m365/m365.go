// Package m365 retrieves the endpoints Microsoft 365 serves from. These are
// not the Azure service tags: Exchange, SharePoint and Teams are published
// separately from the Azure ranges, and it is these an organisation allowlists
// to reach Microsoft 365 itself.
package m365

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "m365"
	FullName  = "Microsoft 365"
	HostType  = "saas"
	SourceURL = "https://learn.microsoft.com/en-us/microsoft-365/enterprise/microsoft-365-ip-web-service"

	// DefaultClientRequestID identifies the caller to the endpoints web
	// service, which requires a GUID on every request and uses it to track
	// which version a caller last saw. Swap it, by setting DownloadURL, if you
	// want your own requests accounted for separately.
	DefaultClientRequestID = "b10c5ed1-bad1-445f-b386-b919946339a7"

	// WorldwideURL returns the endpoints of the worldwide commercial instance,
	// which is the one nearly every tenant is on.
	WorldwideURL = "https://endpoints.office.com/endpoints/worldwide?clientrequestid=" + DefaultClientRequestID
	// ChinaURL returns the endpoints of the instance operated by 21Vianet.
	ChinaURL = "https://endpoints.office.com/endpoints/china?clientrequestid=" + DefaultClientRequestID
	// USGovDoDURL returns the endpoints of the US Government DoD instance.
	USGovDoDURL = "https://endpoints.office.com/endpoints/usgovdod?clientrequestid=" + DefaultClientRequestID
	// USGovGCCHighURL returns the endpoints of the US Government GCC High
	// instance. Set any of these as the DownloadURL to fetch that instance.
	USGovGCCHighURL = "https://endpoints.office.com/endpoints/usgovgcchigh?clientrequestid=" + DefaultClientRequestID

	DownloadURL = WorldwideURL
)

type M365 struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() M365 {
	return M365{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type RawEndpoint struct {
	ID                     int      `json:"id"`
	ServiceArea            string   `json:"serviceArea"`
	ServiceAreaDisplayName string   `json:"serviceAreaDisplayName"`
	URLs                   []string `json:"urls"`
	IPs                    []string `json:"ips"`
	TCPPorts               string   `json:"tcpPorts"`
	UDPPorts               string   `json:"udpPorts"`
	ExpressRoute           bool     `json:"expressRoute"`
	Category               string   `json:"category"`
	Required               bool     `json:"required"`
	Notes                  string   `json:"notes"`
}

// Endpoint keeps the category and the required flag alongside the prefixes,
// as those decide how an endpoint set is treated: Optimize and Allow carry the
// traffic that must not be proxied or inspected, Default the rest.
type Endpoint struct {
	ID                     int            `json:"id"                        yaml:"id"`
	ServiceArea            string         `json:"service_area"              yaml:"service_area"`
	ServiceAreaDisplayName string         `json:"service_area_display_name" yaml:"service_area_display_name"`
	URLs                   []string       `json:"urls,omitempty"            yaml:"urls,omitempty"`
	IPv4Prefixes           []netip.Prefix `json:"ipv4_prefixes"             yaml:"ipv4_prefixes"`
	IPv6Prefixes           []netip.Prefix `json:"ipv6_prefixes"             yaml:"ipv6_prefixes"`
	TCPPorts               string         `json:"tcp_ports,omitempty"       yaml:"tcp_ports,omitempty"`
	UDPPorts               string         `json:"udp_ports,omitempty"       yaml:"udp_ports,omitempty"`
	ExpressRoute           bool           `json:"express_route"             yaml:"express_route"`
	Category               string         `json:"category"                  yaml:"category"`
	Required               bool           `json:"required"                  yaml:"required"`
	Notes                  string         `json:"notes,omitempty"           yaml:"notes,omitempty"`
}

type Doc struct {
	Endpoints []Endpoint `json:"endpoints" yaml:"endpoints"`
}

func (m *M365) FetchData() ([]byte, http.Header, int, error) {
	if m.DownloadURL == "" {
		m.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(m.Client, m.DownloadURL, http.MethodGet, nil, nil, m.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download microsoft 365 endpoints from %s. http status code: %d", m.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (m *M365) Fetch() (Doc, error) {
	data, _, _, err := m.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

// ProcessData drops the endpoint sets published as URLs alone, as they carry no
// addresses to fetch. The document returned by FetchData still holds them.
func ProcessData(data []byte) (Doc, error) {
	var raw []RawEndpoint
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	endpoints := make([]Endpoint, 0, len(raw))

	for _, entry := range raw {
		ipv4, ipv6 := splitFamilies(entry)
		if len(ipv4) == 0 && len(ipv6) == 0 {
			continue
		}

		endpoints = append(endpoints, Endpoint{
			ID:                     entry.ID,
			ServiceArea:            entry.ServiceArea,
			ServiceAreaDisplayName: entry.ServiceAreaDisplayName,
			URLs:                   entry.URLs,
			IPv4Prefixes:           ipv4,
			IPv6Prefixes:           ipv6,
			TCPPorts:               entry.TCPPorts,
			UDPPorts:               entry.UDPPorts,
			ExpressRoute:           entry.ExpressRoute,
			Category:               entry.Category,
			Required:               entry.Required,
			Notes:                  entry.Notes,
		})
	}

	if len(endpoints) == 0 {
		return Doc{}, nil
	}

	return Doc{Endpoints: endpoints}, nil
}

// splitFamilies sorts an endpoint set's addresses into families, logging and
// skipping any entry that cannot be parsed.
func splitFamilies(entry RawEndpoint) ([]netip.Prefix, []netip.Prefix) {
	var ipv4, ipv6 []netip.Prefix

	source := fmt.Sprintf("%s endpoint set %d", ShortName, entry.ID)

	for _, prefix := range iplist.CastPrefixes(source, entry.IPs) {
		if prefix.Addr().Is4() {
			ipv4 = append(ipv4, prefix)

			continue
		}

		ipv6 = append(ipv6, prefix)
	}

	return ipv4, ipv6
}
