// Package ipsum retrieves stamparm's IPsum consensus threat list.
package ipsum

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
	ShortName = "ipsum"
	FullName  = "IPsum"
	HostType  = "threat"
	SourceURL = "https://github.com/stamparm/ipsum"
	// DownloadURL returns addresses that appear on at least three of the
	// blacklists IPsum aggregates, as a newline separated list of bare IPv4
	// addresses. Higher levels (fewer hits required) are available under the
	// same path with a different filename.
	DownloadURL = "https://raw.githubusercontent.com/stamparm/ipsum/master/levels/3.txt"
)

type IPsum struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() IPsum {
	return IPsum{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (i *IPsum) FetchData() ([]byte, http.Header, int, error) {
	if i.DownloadURL == "" {
		i.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(i.Client, i.DownloadURL, http.MethodGet, nil, nil, i.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download ipsum list from %s. http status code: %d", i.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (i *IPsum) Fetch() (Doc, error) {
	data, _, _, err := i.FetchData()
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
