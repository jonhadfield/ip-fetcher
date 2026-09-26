// Package hetrixtools retrieves HetrixTools uptime monitoring probe addresses.
package hetrixtools

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
	ShortName = "hetrixtools"
	FullName  = "HetrixTools"
	HostType  = "monitoring"
	SourceURL = "https://docs.hetrixtools.com/uptime-monitoring-ip-addresses/"
	// DownloadURL returns the bare IPv4 addresses of HetrixTools uptime probes.
	DownloadURL = "https://hetrixtools.com/resources/uptime-monitor-only-ips.txt"
)

type HetrixTools struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() HetrixTools {
	return HetrixTools{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (h *HetrixTools) FetchData() ([]byte, http.Header, int, error) {
	if h.DownloadURL == "" {
		h.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(h.Client, h.DownloadURL, http.MethodGet, nil, nil, h.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download hetrixtools addresses from %s. http status code: %d", h.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (h *HetrixTools) Fetch() (Doc, error) {
	data, _, _, err := h.FetchData()
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
