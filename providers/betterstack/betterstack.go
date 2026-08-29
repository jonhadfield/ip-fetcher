package betterstack

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
	ShortName = "betterstack"
	FullName  = "Better Stack"
	HostType  = "monitoring"
	SourceURL = "https://betterstack.com/docs/uptime/ip-addresses/"
	// DownloadURL returns the addresses Better Stack's uptime checks run from,
	// as a newline separated list carrying both address families.
	DownloadURL = "https://uptime.betterstack.com/ips.txt"
)

type BetterStack struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() BetterStack {
	return BetterStack{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (b *BetterStack) FetchData() ([]byte, http.Header, int, error) {
	if b.DownloadURL == "" {
		b.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(b.Client, b.DownloadURL, http.MethodGet, nil, nil, b.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download betterstack list from %s. http status code: %d", b.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (b *BetterStack) Fetch() (Doc, error) {
	data, _, _, err := b.FetchData()
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
