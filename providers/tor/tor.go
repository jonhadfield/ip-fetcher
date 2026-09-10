// Package tor retrieves the addresses of the Tor network's exit nodes, being
// the addresses traffic leaving the network arrives from.
package tor

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
	ShortName = "tor"
	FullName  = "Tor Exit Nodes"
	HostType  = "anonymiser"
	SourceURL = "https://check.torproject.org/"
	// DownloadURL returns the addresses of the exits the Tor Project's own
	// checker has confirmed, as a newline separated list of bare addresses.
	// The list holds only exits, not the relays traffic passes through
	// earlier, as those never appear as the source of a connection.
	DownloadURL = "https://check.torproject.org/torbulkexitlist"
)

type Tor struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Tor {
	return Tor{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (t *Tor) FetchData() ([]byte, http.Header, int, error) {
	if t.DownloadURL == "" {
		t.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(t.Client, t.DownloadURL, http.MethodGet, nil, nil, t.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download tor exit list from %s. http status code: %d", t.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (t *Tor) Fetch() (Doc, error) {
	data, _, _, err := t.FetchData()
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
