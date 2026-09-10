// Package feodo retrieves abuse.ch's Feodo Tracker blocklist, being the
// addresses of the botnet command and control servers it tracks.
package feodo

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
	ShortName = "feodo"
	FullName  = "abuse.ch Feodo Tracker"
	HostType  = "threat"
	SourceURL = "https://feodotracker.abuse.ch/blocklist/"
	// DownloadURL returns every command and control server Feodo Tracker
	// currently holds, as a newline separated list of bare addresses behind a
	// commented header. The list is regenerated every five minutes.
	DownloadURL = "https://feodotracker.abuse.ch/downloads/ipblocklist.txt"
	// RecommendedDownloadURL returns the subset abuse.ch recommends blocking
	// on, which drops the entries most likely to be false positives. Set it as
	// the DownloadURL to fetch that list instead.
	RecommendedDownloadURL = "https://feodotracker.abuse.ch/downloads/ipblocklist_recommended.txt"
)

type Feodo struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Feodo {
	return Feodo{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (f *Feodo) FetchData() ([]byte, http.Header, int, error) {
	if f.DownloadURL == "" {
		f.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(f.Client, f.DownloadURL, http.MethodGet, nil, nil, f.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download feodo tracker list from %s. http status code: %d", f.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (f *Feodo) Fetch() (Doc, error) {
	data, _, _, err := f.FetchData()
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
