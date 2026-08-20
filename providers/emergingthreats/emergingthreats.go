package emergingthreats

import (
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "emergingthreats"
	FullName  = "Emerging Threats Compromised IPs"
	HostType  = "threat"
	SourceURL = "https://rules.emergingthreats.net/blockrules/"
	// DownloadURL returns the compromised host list as a newline separated list
	// of bare IPv4 addresses.
	DownloadURL = "https://rules.emergingthreats.net/blockrules/compromised-ips.txt"
)

type EmergingThreats struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() EmergingThreats {
	return EmergingThreats{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (e *EmergingThreats) FetchData() ([]byte, http.Header, int, error) {
	if e.DownloadURL == "" {
		e.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(e.Client, e.DownloadURL, http.MethodGet, nil, nil, e.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download emergingthreats list from %s. http status code: %d", e.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (e *EmergingThreats) Fetch() (Doc, error) {
	data, _, _, err := e.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	ipv4, ipv6, err := iplist.Parse(ShortName, data)
	if err != nil {
		return Doc{}, err
	}

	return Doc{IPv4Prefixes: ipv4, IPv6Prefixes: ipv6}, nil
}
